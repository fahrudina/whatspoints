-- Test script to verify the registration fix
-- Run this to check if the registration functionality works properly

-- Test 1: Check if the members table exists and has the correct structure
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = 'members' 
ORDER BY ordinal_position;

-- Test 2: Check if the points table exists and has the correct structure
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = 'points' 
ORDER BY ordinal_position;

-- Test 3: Test insert with PostgreSQL syntax (this should work now)
DO $$
DECLARE
    test_member_id INT;
BEGIN
    -- Insert a test member
    INSERT INTO members (name, address, phone_number, created_at, updated_at) 
    VALUES ('Test User', 'Test Address', '1234567890', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP) 
    RETURNING member_id INTO test_member_id;
    
    -- Insert corresponding points record
    INSERT INTO points (member_id, accumulated_points, current_points, created_at, updated_at) 
    VALUES (test_member_id, 0, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
    
    -- Clean up the test data
    DELETE FROM points WHERE member_id = test_member_id;
    DELETE FROM members WHERE member_id = test_member_id;
    
    RAISE NOTICE 'Test passed: Registration SQL syntax is correct!';
END $$;
