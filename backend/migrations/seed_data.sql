BEGIN;

-- sample products
INSERT INTO products (sku, name, price, stock) VALUES
('SKU-001', 'Sugar Zucchero', 250000, 95),
('SKU-002', 'Cookies Cappucino', 300000, 125),
('SKU-003', 'Notte Stelle', 450000, 118),
('SKU-004', 'Bianco Ariston', 220000, 150),
('SKU-005', 'Statuario Venatino', 300000, 153),
('SKU-006', 'Milky Latte', 500000, 134),
('SKU-007', 'Pompeii Grigio', 325000, 131),
('SKU-008', 'Sedimen Tario', 175000, 151),
('SKU-009', 'Macabo Grey', 150000, 142),
('SKU-010', 'Mocha Cioccolatta', 135000, 100),
('SKU-011', 'Vulcan Ash', 275000, 180),
('SKU-012', 'Portoro Nero', 425000, 130);

-- sample stock movements as inbound stock
INSERT INTO stock_movements (product_id, type, qty, note)
SELECT
    p.id,
    'in',
    p.stock + CASE p.sku
        WHEN 'SKU-001' THEN 10
        WHEN 'SKU-003' THEN 13
        WHEN 'SKU-005' THEN 16
        WHEN 'SKU-007' THEN 19
        ELSE 0
    END,
    'Initial stock inbound'
FROM products p;

-- sample sales orders
INSERT INTO sales_orders (customer, total)
VALUES
    ('PT Astro Design', 12750000),
    ('PT Semesta Design', 7875000);

-- sample order items
INSERT INTO sales_order_items
    (order_id, sku, qty, unit_price)
VALUES
    (1, 'SKU-001', 15, 250000),
    (1, 'SKU-003', 20, 450000),
    (2, 'SKU-005', 10, 300000),
    (2, 'SKU-007', 15, 325000);

-- sample stock movements from sample orders
INSERT INTO stock_movements (product_id, type, qty, note)
SELECT
    p.id,
    'out',
    sold.qty,
    'Sample sales order'
FROM (
    VALUES
        ('SKU-001', 15),
        ('SKU-003', 20),
        ('SKU-005', 10),
        ('SKU-007', 15)
) AS sold(sku, qty)
JOIN products p ON p.sku = sold.sku;

COMMIT;