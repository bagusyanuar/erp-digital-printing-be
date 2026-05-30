-- Create finishings table
CREATE TABLE IF NOT EXISTS finishings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    price DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create orders table
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_number VARCHAR(100) UNIQUE NOT NULL,
    invoice_number VARCHAR(100) UNIQUE,
    reseller_id UUID REFERENCES resellers(id) ON DELETE RESTRICT,
    designer_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    cashier_id UUID REFERENCES users(id) ON DELETE RESTRICT,
    customer_name VARCHAR(255),
    customer_phone VARCHAR(20),
    status VARCHAR(50) NOT NULL DEFAULT 'DRAFT',
    payment_status VARCHAR(50) NOT NULL DEFAULT 'UNPAID',
    notes TEXT,
    total_additional_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    total_product_price DECIMAL(15,2) NOT NULL DEFAULT 0,
    grand_total DECIMAL(15,2) NOT NULL DEFAULT 0,
    amount_paid DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create order_items table
CREATE TABLE IF NOT EXISTS order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    product_variant_id UUID NOT NULL REFERENCES product_variants(id) ON DELETE RESTRICT,
    uom VARCHAR(50) NOT NULL,
    length_cm DECIMAL(10,2),
    width_cm DECIMAL(10,2),
    quantity INT NOT NULL,
    design_file_url TEXT,
    production_notes TEXT,
    price_per_unit DECIMAL(15,2) NOT NULL DEFAULT 0,
    additional_cost DECIMAL(15,2) NOT NULL DEFAULT 0,
    subtotal DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- Create order_item_finishings pivot table
CREATE TABLE IF NOT EXISTS order_item_finishings (
    order_item_id UUID NOT NULL REFERENCES order_items(id) ON DELETE CASCADE,
    finishing_id UUID NOT NULL REFERENCES finishings(id) ON DELETE RESTRICT,
    PRIMARY KEY (order_item_id, finishing_id)
);
