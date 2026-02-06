CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    address TEXT,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    budget NUMERIC(15,2),
    status TEXT DEFAULT 'planning',
    characteristics JSONB, -- Добавлено
    cost_estimates JSONB, -- Добавлено
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    duration_days INTEGER NOT NULL,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    dependencies JSONB DEFAULT '[]'::JSONB,
    actual_progress NUMERIC(3,2) DEFAULT 0,
    status TEXT DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE resources (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    quantity NUMERIC(10,2) NOT NULL,
    unit TEXT NOT NULL,
    unit_cost NUMERIC(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE documents (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    path TEXT NOT NULL,
    hash TEXT NOT NULL,
    stage TEXT NOT NULL,      -- 'P' or 'R'
    doc_type TEXT NOT NULL,   -- 'II' or other
    version TEXT NOT NULL,
    parsed_data JSONB, -- Добавлено для парсинга
    changed_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_tasks_project_id ON tasks(project_id);
CREATE INDEX idx_resources_task_id ON resources(task_id);
CREATE INDEX idx_documents_project_id ON documents(project_id);
CREATE INDEX idx_documents_hash ON documents(hash);
