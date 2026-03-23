// web/static/js/auth.js

const API_BASE = '/api/v1';
const TOKEN_KEY = 'ai_construction_token';

// Проверка авторизации
export function isAuthenticated() {
    return localStorage.getItem(TOKEN_KEY) !== null;
}

// Получение токена
export function getToken() {
    return localStorage.getItem(TOKEN_KEY);
}

// Установка токена
export function setToken(token) {
    localStorage.setItem(TOKEN_KEY, token);
}

// Выход
export function logout() {
    localStorage.removeItem(TOKEN_KEY);
    window.location.href = '/login';
}

// Авторизованный fetch
export async function apiFetch(endpoint, options = {}) {
    const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
    };
    
    const token = getToken();
    if (token) {
        headers['Authorization'] = `Bearer ${token}`;
    }
    
    const response = await fetch(`${API_BASE}${endpoint}`, {
        ...options,
        headers,
    });
    
    if (response.status === 401) {
        logout();
        throw new Error('Unauthorized');
    }
    
    if (!response.ok) {
        const error = await response.json().catch(() => ({}));
        throw new Error(error.error || 'Request failed');
    }
    
    return response.json();
}

// Логин
export async function login(username, password) {
    // TODO: Реализовать эндпоинт /auth/login на бэкенде
    // Пока заглушка для демо:
    if (username === 'admin' && password === 'admin') {
        const token = btoa(JSON.stringify({ 
            sub: username, 
            exp: Date.now() + 24*60*60*1000 
        }));
        setToken(token);
        return true;
    }
    throw new Error('Неверные учётные данные');
}

// Проверка авторизации при загрузке страницы
export function requireAuth() {
    if (!isAuthenticated()) {
        window.location.href = '/login';
        return false;
    }
    return true;
}
