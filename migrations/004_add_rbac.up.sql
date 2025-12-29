BEGIN;


CREATE TABLE roles (
                       id   SERIAL PRIMARY KEY,
                       code TEXT NOT NULL UNIQUE,   -- admin, librarian, reader
                       name TEXT NOT NULL
);


CREATE TABLE permissions (
                             id   SERIAL PRIMARY KEY,
                             code TEXT NOT NULL UNIQUE,   -- books.view, books.create, books.delete
                             name TEXT NOT NULL
);

CREATE TABLE role_permissions (
                                  role_id       INT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                                  permission_id INT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
                                  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
                            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            role_id INT  NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
                            PRIMARY KEY (user_id, role_id)
);

-- Роли
INSERT INTO roles (code, name) VALUES
                                   ('admin',      'Администратор'),
                                   ('librarian',  'Библиотекарь'),
                                   ('reader',     'Читатель')
    ON CONFLICT (code) DO NOTHING;

-- Права
INSERT INTO permissions (code, name) VALUES
                                         ('books.view',   'Просмотр детальной информации о книге'),
                                         ('books.create', 'Добавление книг'),
                                         ('books.update', 'Редактирование книг'),
                                         ('books.delete', 'Удаление книг')
    ON CONFLICT (code) DO NOTHING;

-- Администратор — все права
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         CROSS JOIN permissions p
WHERE r.code = 'admin'
    ON CONFLICT DO NOTHING;

-- Библиотекарь — без удаления
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON p.code IN ('books.view', 'books.create', 'books.update')
WHERE r.code = 'librarian'
    ON CONFLICT DO NOTHING;

-- Читатель — только просмотр
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
         JOIN permissions p ON p.code = 'books.view'
WHERE r.code = 'reader'
    ON CONFLICT DO NOTHING;

COMMIT;
