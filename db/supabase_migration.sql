-- Supabase/PostgreSQL migration file
-- Adjust table names and columns as needed for your app

CREATE TABLE IF NOT EXISTS members (
    id SERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL UNIQUE,
    name VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS images (
    id SERIAL PRIMARY KEY,
    member_id INTEGER REFERENCES members(id),
    s3_url TEXT NOT NULL,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS points (
    id SERIAL PRIMARY KEY,
    member_id INTEGER REFERENCES members(id),
    points INTEGER NOT NULL,
    receipt_image_id INTEGER REFERENCES images(id),
    earned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    member_id INTEGER REFERENCES members(id),
    points INTEGER NOT NULL,
    type VARCHAR(20) NOT NULL, -- earn or redeem
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
