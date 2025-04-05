-- Table: MEMBER
CREATE TABLE MEMBER (
    member_id INT PRIMARY KEY,
    phone_number VARCHAR(20) UNIQUE, -- "Unique identifier"
    name VARCHAR(100),
    address TEXT,
    created_at DATETIME,
    updated_at DATETIME
);

-- Table: POINT
CREATE TABLE POINT (
    point_id INT PRIMARY KEY,
    member_id INT,
    accumulated_points INT,
    current_points INT,
    created_at DATETIME,
    updated_at DATETIME,
    FOREIGN KEY (member_id) REFERENCES MEMBER(member_id)
);

-- Table: RECEIPT
CREATE TABLE RECEIPT (
    receipt_id INT PRIMARY KEY,
    member_id INT,
    receipt_image TEXT, -- You can change to BLOB if storing binary
    total_kg DECIMAL(10, 2),
    total_unit INT,
    total_price DECIMAL(10, 2),
    points_earned INT,
    receipt_date DATETIME,
    created_at DATETIME,
    updated_at DATETIME,
    FOREIGN KEY (member_id) REFERENCES MEMBER(member_id)
);

-- Table: POINT_TRANSACTION
CREATE TABLE POINT_TRANSACTION (
    transaction_id INT PRIMARY KEY,
    point_id INT,
    receipt_id INT NULL,
    points_changed INT,
    transaction_type VARCHAR(20), -- e.g., 'EARN', 'REDEEM', 'EXPIRE'
    transaction_date DATETIME,
    notes TEXT,
    FOREIGN KEY (point_id) REFERENCES POINT(point_id),
    FOREIGN KEY (receipt_id) REFERENCES RECEIPT(receipt_id)
);
