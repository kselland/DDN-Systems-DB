CREATE TYPE product_type AS ENUM ('cabinet', 'accessory');

CREATE TABLE colors (
    id SERIAL PRIMARY KEY,
    hex_code text NOT NULL CHECK (hex_code ~ '^#([A-F0-9]{6})$'),
    name text NOT NULL,
    product_type product_type NOT NULL
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
    color_id integer NOT NULL,
    external_id text NOT NULL,
    FOREIGN KEY (color_id) REFERENCES colors (id)
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
