-- Table: members
CREATE TABLE members (
    member_id INT PRIMARY KEY AUTO_INCREMENT,
    phone_number VARCHAR(20) UNIQUE, -- "Unique identifier"
    name VARCHAR(100),
    address TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- Table: points
CREATE TABLE points (
    point_id INT PRIMARY KEY AUTO_INCREMENT,
    member_id INT,
    accumulated_points INT,
    current_points INT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: receipts
CREATE TABLE receipts (
    receipt_id INT PRIMARY KEY AUTO_INCREMENT,
    member_id INT,
    receipt_image TEXT, -- URL of the receipt image
    total_kg DECIMAL(10, 2),
    total_unit INT,
    total_price DECIMAL(10, 2),
    points_earned INT,
    receipt_date DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);

-- Table: point_transactions
CREATE TABLE point_transactions (
    transaction_id INT PRIMARY KEY AUTO_INCREMENT,
    point_id INT,
    receipt_id INT NULL,
    points_changed INT,
    transaction_type VARCHAR(20), -- e.g., 'EARN', 'REDEEM', 'EXPIRE'
    transaction_date DATETIME,
    notes TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (point_id) REFERENCES points(point_id),
    FOREIGN KEY (receipt_id) REFERENCES receipts(receipt_id)
);

-- Table: images
CREATE TABLE images (
    image_id INT PRIMARY KEY AUTO_INCREMENT,
    member_id INT,
    image_url TEXT, 
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (member_id) REFERENCES members(member_id)
);