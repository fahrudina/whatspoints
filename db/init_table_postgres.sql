-- PostgreSQL/Supabase compatible schema migration

-- Table: members
CREATE TABLE IF NOT EXISTS members (
    member_id SERIAL PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE,
    name VARCHAR(100),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: points
CREATE TABLE IF NOT EXISTS points (
    point_id SERIAL PRIMARY KEY,
    member_id INTEGER UNIQUE,
    accumulated_points INTEGER,
    current_points INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: receipts
CREATE TABLE IF NOT EXISTS receipts (
    receipt_id SERIAL PRIMARY KEY,
    member_id INTEGER,
    receipt_image TEXT,
    total_kg NUMERIC(10, 2),
    total_unit INTEGER,
    total_price NUMERIC(10, 2),
    points_earned INTEGER,
    receipt_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: point_transactions
CREATE TABLE IF NOT EXISTS point_transactions (
    transaction_id SERIAL PRIMARY KEY,
    point_id INTEGER,
    receipt_id INTEGER,
    points_changed INTEGER,
    transaction_type VARCHAR(20),
    transaction_date TIMESTAMP,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (point_id) REFERENCES points(point_id),
    FOREIGN KEY (receipt_id) REFERENCES receipts(receipt_id)
);

-- Table: images
CREATE TABLE IF NOT EXISTS images (
    image_id SERIAL PRIMARY KEY,
    member_id INTEGER,
    image_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: orders
CREATE TABLE IF NOT EXISTS orders (
    order_id SERIAL PRIMARY KEY,
    member_id INTEGER,
    total_price NUMERIC(10, 2),
    order_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: items
CREATE TABLE IF NOT EXISTS items (
    item_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_per_unit NUMERIC(10, 2) DEFAULT 0.00,
    price_per_kilo NUMERIC(10, 2) DEFAULT 0.00,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Table: order_items
CREATE TABLE IF NOT EXISTS order_items (
    order_item_id SERIAL PRIMARY KEY,
    order_id INTEGER,
    item_id INTEGER,
    total_kilo NUMERIC(10, 2) DEFAULT 0.00,
    total_unit INTEGER DEFAULT 0,
    price NUMERIC(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(order_id),
    FOREIGN KEY (item_id) REFERENCES items(item_id)
);
