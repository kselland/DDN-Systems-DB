-- APPLICATION SPECIFIC TABLES
CREATE TYPE product_type AS ENUM ('cabinet', 'accessory');

CREATE TABLE colors (
    name text PRIMARY KEY,
    hex_code text NOT NULL CHECK (hex_code ~ '^#([A-F0-9]{6})$')
);

CREATE TABLE color_product_types (
    id SERIAL PRIMARY KEY,
    product_type product_type NOT NULL,
    color_name text NOT NULL,
    FOREIGN KEY  (color_name) REFERENCES colors (name),
    UNIQUE (product_type, color_name)
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name text NOT NULL,
    product_type product_type NOT NULL,
    length integer NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL,
    active boolean NOT NULL,
    price_cents integer NOT NULL,
    color_name text NOT NULL,
    FOREIGN KEY (color_name, product_type) REFERENCES color_product_types (color_name, product_type)
);

CREATE TABLE storage_locations (
    id SERIAL PRIMARY KEY,
    bin text NOT NULL,
    length integer NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL
);

CREATE TABLE inventory_items (
    id SERIAL PRIMARY KEY,
    product_id integer NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products (id),
    quantity integer NOT NULL,
    batch_number integer NOT NULL,
    storage_location_id integer NOT NULL,
    FOREIGN KEY (storage_location_id) REFERENCES storage_locations (id)
);

-- AUTH RELATED TABLES
CREATE TABLE sessions (
    id SERIAL PRIMARY KEY,
    csrf_token text NOT NULL,
    user_id integer NOT NULL,
    session_key_digest bytea NOT NULL
);

CREATE TYPE user_role AS ENUM ('superadmin', 'admin', 'manager', 'viewer', 'builder', 'driver');

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    password_digest bytea NOT NULL,
    email text NOT NULL UNIQUE,
    name text NOT NULL,
    role user_role NOT NULL
);
