CREATE TABLE sellers (
    id VARCHAR(36) PRIMARY KEY,
    -- Assuming UUIDs are used for IDs
    brand_id VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL,
    address VARCHAR(255) NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    country VARCHAR(100) NOT NULL DEFAULT 'AUS',
    postcode VARCHAR(20) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(30) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    -- Precision for latitude
    longitude DECIMAL(11, 8) NOT NULL,
    -- Precision for longitude
    last_updated_by VARCHAR(36) NOT NULL,
    -- Assuming User ID is also a UUID or similar
    last_update_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- Optional: Add an index for frequently queried fields like email or brand_id
CREATE INDEX idx_sellers_email ON sellers(email);
CREATE INDEX idx_sellers_brand_id ON sellers(brand_id);
CREATE INDEX idx_sellers_status ON sellers(status);
-- Optional: Add a unique constraint on email if emails must be unique
-- ALTER TABLE sellers ADD CONSTRAINT uq_sellers_email UNIQUE (email);
COMMENT ON COLUMN sellers.id IS 'Unique identifier for the seller (e.g., UUID)';
COMMENT ON COLUMN sellers.brand_id IS 'Identifier for the brand associated with the seller';
COMMENT ON COLUMN sellers.status IS 'Current status of the seller (e.g., ACTIVE, PENDING)';
COMMENT ON COLUMN sellers.address IS 'Street address of the seller';
COMMENT ON COLUMN sellers.city IS 'City of the seller';
COMMENT ON COLUMN sellers.state IS 'State or province of the seller';
COMMENT ON COLUMN sellers.country IS 'Country of the seller, defaults to AUS';
COMMENT ON COLUMN sellers.postcode IS 'Postal code of the seller';
COMMENT ON COLUMN sellers.email IS 'Contact email address of the seller';
COMMENT ON COLUMN sellers.phone_number IS 'Contact phone number of the seller';
COMMENT ON COLUMN sellers.latitude IS 'Geographical latitude of the seller';
COMMENT ON COLUMN sellers.longitude IS 'Geographical longitude of the seller';
COMMENT ON COLUMN sellers.last_updated_by IS 'User ID of the person who last updated the record';
COMMENT ON COLUMN sellers.last_update_time IS 'Timestamp of when the record was last updated';