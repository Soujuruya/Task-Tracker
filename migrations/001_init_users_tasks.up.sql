CREATE TABLE users (
                       id UUID PRIMARY KEY,
                       username TEXT NOT NULL UNIQUE,
                       password_hash TEXT NOT NULL
);

CREATE TABLE tasks (
                       id UUID PRIMARY KEY,
                       user_id UUID NOT NULL,
                       title TEXT NOT NULL,
                       description TEXT NOT NULL,
                       progress_status TEXT NOT NULL,
                       created_at TIMESTAMPTZ NOT NULL,
                       CONSTRAINT fk_tasks_user
                           FOREIGN KEY (user_id) REFERENCES users(id),
                       CONSTRAINT tasks_progress_status
                           CHECK (progress_status IN ('todo', 'in_progress', 'done', 'blocked'))
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);