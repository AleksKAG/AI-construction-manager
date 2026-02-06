-- Включаем поддержку UUID
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 1. Организации
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('заказчик', 'генподрядчик', 'проектировщик', 'экспертиза', 'поставщик')),
    contact_person TEXT,
    phone TEXT,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Специалисты
CREATE TABLE specialists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name TEXT NOT NULL,
    role TEXT NOT NULL, -- архитектор, инженер, прораб и т.д.
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    phone TEXT,
    email TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 3. Объекты строительства
CREATE TABLE construction_objects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    address TEXT,
    type TEXT NOT NULL CHECK (type IN ('новое строительство', 'реконструкция', 'капитальный ремонт')),
    status TEXT NOT NULL DEFAULT 'проектирование' CHECK (status IN ('проектирование', 'строительство', 'сдан', 'приостановлен')),
    start_date DATE,
    end_date DATE,
    client_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    characteristics JSONB, -- Добавлено для хранения характеристик (map[string]string)
    cost_estimates JSONB, -- Добавлено для оценок (map[string]float64)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Этапы проекта (ПИР / СМР)
CREATE TABLE project_phases (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id UUID NOT NULL REFERENCES construction_objects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    code TEXT,
    stage TEXT NOT NULL CHECK (stage IN ('П', 'Р')), -- П = проект, Р = рабочая документация
    start_date DATE,
    end_date DATE,
    status TEXT NOT NULL DEFAULT 'запланирован' CHECK (status IN ('запланирован', 'в работе', 'на согласовании', 'согласован', 'завершён')),
    responsible_specialist_id UUID REFERENCES specialists(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 5. Состав документации (П и РД)
CREATE TABLE documentation_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phase_id UUID NOT NULL REFERENCES project_phases(id) ON DELETE CASCADE,
    code TEXT NOT NULL, -- АР, КР, ОВ, ЭО, ПЗ и т.д.
    title TEXT NOT NULL,
    stage TEXT NOT NULL CHECK (stage IN ('П', 'Р')),
    required BOOLEAN DEFAULT true,
    normative_reference TEXT, -- ГОСТ, СП, внутреннее ТЗ
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 6. Документы (файлы)
CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    item_id UUID NOT NULL REFERENCES documentation_items(id) ON DELETE CASCADE,
    version INT NOT NULL DEFAULT 1,
    file_path TEXT NOT NULL, -- путь в MinIO/S3
    uploaded_by UUID REFERENCES specialists(id) ON DELETE SET NULL,
    uploaded_at TIMESTAMPTZ DEFAULT NOW(),
    status TEXT NOT NULL DEFAULT 'черновик' CHECK (status IN ('черновик', 'на согласовании', 'согласован', 'отклонён')),
    comments TEXT,
    parsed_content JSONB, -- Добавлено для хранения распарсенного контента (из upload)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 7. Виды работ (СМР)
CREATE TABLE work_types (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    code TEXT,
    category TEXT NOT NULL CHECK (category IN ('общестроительные', 'инженерные', 'благоустройство', 'геодезия', 'лаборатория')),
    unit TEXT NOT NULL, -- м², м³, шт, тонн
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 8. Сметные нормативы
CREATE TABLE cost_norms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    work_type_id UUID NOT NULL REFERENCES work_types(id) ON DELETE CASCADE,
    region TEXT NOT NULL DEFAULT 'Москва',
    base_cost NUMERIC(12,2) NOT NULL,
    inflation_index NUMERIC(5,4) DEFAULT 1.0000,
    valid_from DATE NOT NULL,
    valid_to DATE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 9. Материалы и оборудование
CREATE TABLE materials (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('материал', 'оборудование')),
    unit TEXT NOT NULL,
    brand TEXT,
    certified BOOLEAN DEFAULT true,
    supplier_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 10. Связь материалов с работами
CREATE TABLE work_materials (
    work_type_id UUID NOT NULL REFERENCES work_types(id) ON DELETE CASCADE,
    material_id UUID NOT NULL REFERENCES materials(id) ON DELETE CASCADE,
    quantity_per_unit NUMERIC(10,3) NOT NULL,
    PRIMARY KEY (work_type_id, material_id)
);

-- 11. График работ (СМР)
CREATE TABLE schedule_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id UUID NOT NULL REFERENCES construction_objects(id) ON DELETE CASCADE,
    work_type_id UUID NOT NULL REFERENCES work_types(id) ON DELETE RESTRICT,
    start_date DATE,
    end_date DATE,
    dependencies JSONB, -- массив UUID зависимых задач
    assigned_to UUID REFERENCES organizations(id) ON DELETE SET NULL, -- подрядчик
    status TEXT NOT NULL DEFAULT 'не начато' CHECK (status IN ('не начато', 'в работе', 'задержка', 'завершено')),
    duration INT, -- Добавлено для Gantt
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 12. Риски
CREATE TABLE risks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    object_id UUID NOT NULL REFERENCES construction_objects(id) ON DELETE CASCADE,
    description TEXT NOT NULL,
    type TEXT NOT NULL CHECK (type IN ('погода', 'поставки', 'согласования', 'персонал', 'финансы')),
    probability NUMERIC(3,2) CHECK (probability BETWEEN 0 AND 1),
    impact TEXT CHECK (impact IN ('низкий', 'средний', 'высокий')),
    mitigation_plan TEXT,
    ml_prediction JSONB, -- Добавлено для хранения ML-прогнозов (Gonum)
    detected_at TIMESTAMPTZ DEFAULT NOW()
);

-- 13. Согласования
CREATE TABLE approvals (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    document_id UUID NOT NULL REFERENCES documents(id) ON DELETE CASCADE,
    approver_id UUID REFERENCES specialists(id) ON DELETE SET NULL,
    approval_type TEXT NOT NULL, -- МЧС, Мосгосстройнадзор, внутреннее
    status TEXT NOT NULL DEFAULT 'ожидает' CHECK (status IN ('ожидает', 'одобрено', 'отклонено')),
    comment TEXT,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 14. История изменений (аудит)
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name TEXT NOT NULL,
    record_id UUID NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('create', 'update', 'delete')),
    changed_by UUID REFERENCES specialists(id) ON DELETE SET NULL,
    changed_at TIMESTAMPTZ DEFAULT NOW(),
    old_values JSONB,
    new_values JSONB
);

-- === ИНДЕКСЫ ===
CREATE INDEX idx_construction_objects_client ON construction_objects(client_id);
CREATE INDEX idx_project_phases_object ON project_phases(object_id);
CREATE INDEX idx_project_phases_status ON project_phases(status);
CREATE INDEX idx_documents_item ON documents(item_id);
CREATE INDEX idx_documents_status ON documents(status);
CREATE INDEX idx_schedule_tasks_object ON schedule_tasks(object_id);
CREATE INDEX idx_schedule_tasks_status ON schedule_tasks(status);
CREATE INDEX idx_risks_object ON risks(object_id);
CREATE INDEX idx_approvals_document ON approvals(document_id);
CREATE INDEX idx_approvals_status ON approvals(status);

-- === ТРИГГЕРЫ ДЛЯ АУДИТА ===
CREATE OR REPLACE FUNCTION audit_construction_objects() RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'DELETE') THEN
        INSERT INTO audit_log (table_name, record_id, action, old_values)
        VALUES (TG_RELNAME, OLD.id, 'delete', row_to_json(OLD));
        RETURN OLD;
    ELSIF (TG_OP = 'UPDATE') THEN
        INSERT INTO audit_log (table_name, record_id, action, old_values, new_values)
        VALUES (TG_RELNAME, NEW.id, 'update', row_to_json(OLD), row_to_json(NEW));
        RETURN NEW;
    ELSIF (TG_OP = 'INSERT') THEN
        INSERT INTO audit_log (table_name, record_id, action, new_values)
        VALUES (TG_RELNAME, NEW.id, 'create', row_to_json(NEW));
        RETURN NEW;
    END IF;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER audit_construction_objects
AFTER INSERT OR UPDATE OR DELETE ON construction_objects
FOR EACH ROW EXECUTE PROCEDURE audit_construction_objects();
