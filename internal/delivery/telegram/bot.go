package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/AleksKAG/ai-construction-manager/internal/domain"
	"github.com/AleksKAG/ai-construction-manager/internal/repository/postgres"
)

type Bot struct {
	api   *tgbotapi.BotAPI
	repo  *postgres.ProjectRepository
	admin int64
}

func NewBot(token string, repo *postgres.ProjectRepository, adminID int64) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	api.Debug = false
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{api: api, repo: repo, admin: adminID}, nil
}

func (b *Bot) Start(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 30

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-updates:
			if update.Message == nil {
				continue
			}
			go b.handleUpdate(update)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	msg := update.Message

	// Базовая безопасность
	if msg.Chat.ID != b.admin && !strings.Contains(msg.Chat.Title, "Стройка") {
		return
	}

	text := strings.ToLower(msg.Text)

	switch {
	case strings.HasPrefix(text, "/start"):
		b.send(msg.Chat.ID, "🏗️ СтройАссистент готов!\n\nКоманды:\n/add_task — добавить задачу\n/status — статус проекта")

	case strings.HasPrefix(text, "/add_task"):
		b.handleAddTask(msg)

	case strings.HasPrefix(text, "/status"):
		b.send(msg.Chat.ID, "📊 Функция /status в разработке")

	default:
		if strings.Contains(text, ":") {
			b.parseNaturalTask(msg)
		} else {
			b.send(msg.Chat.ID, "Неизвестная команда. Используйте /start для списка команд.")
		}
	}
}

func (b *Bot) handleAddTask(msg *tgbotapi.Message) {
	parts := strings.Split(strings.TrimPrefix(msg.Text, "/add_task"), ",")
	if len(parts) < 2 {
		b.send(msg.Chat.ID, "Формат: /add_task Название задачи, 10 дней")
		return
	}

	task := domain.Task{
		Name:         strings.TrimSpace(parts[0]),
		DurationDays: parseInt(strings.TrimSpace(parts[1])),
		Status:       "pending",
		CreatedAt:    time.Now(),
	}

	// TODO: сохранить в БД
	b.send(msg.Chat.ID, fmt.Sprintf("✅ Задача «%s» добавлена на %d дней", task.Name, task.DurationDays))
}

func (b *Bot) parseNaturalTask(msg *tgbotapi.Message) {
	text := strings.ToLower(msg.Text)
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		name := strings.TrimSpace(parts[0])
		details := strings.TrimSpace(parts[1])

		task := domain.Task{
			Name:         capitalize(name),
			DurationDays: extractDays(details),
			Status:       "pending",
			CreatedAt:    time.Now(),
		}

		b.send(msg.Chat.ID, fmt.Sprintf("✅ Распознана задача: %s (%d дней)", task.Name, task.DurationDays))
	}
}

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	b.api.Send(msg)
}

// Вспомогательные функции
func parseInt(s string) int {
	var n int
	fmt.Sscanf(strings.ReplaceAll(s, "дней", ""), "%d", &n)
	return n
}

func extractDays(s string) int {
	var days int
	fmt.Sscanf(s, "%d", &days)
	return days
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(string(s[0])) + s[1:]
}
