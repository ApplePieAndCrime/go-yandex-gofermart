

-- пользователи
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    login TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

ALTER TABLE users ADD COLUMN current_balance DECIMAL(10,2) NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN withdrawn_total DECIMAL(10,2) NOT NULL DEFAULT 0;

-- заказы
CREATE TABLE IF NOT EXISTS orders (
    id SERIAL PRIMARY KEY,
    number TEXT UNIQUE NOT NULL,        -- номер заказа, уникален
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    status TEXT NOT NULL,               -- NEW, PROCESSING, INVALID, PROCESSED
    accrual DECIMAL(10,2),              -- может быть NULL если не начислено
    uploaded_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()  -- для отслеживания обновлений
);

-- списания
CREATE TABLE IF NOT EXISTS withdrawals (
    id SERIAL PRIMARY KEY,
    order_number TEXT NOT NULL,         -- номер заказа, на который списываем
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    sum DECIMAL(10,2) NOT NULL,
    processed_at TIMESTAMP DEFAULT NOW()
);