-- pgview demo database
-- Loaded automatically by postgres:16-alpine on first boot via /docker-entrypoint-initdb.d/
--
-- public schema  →  used by the demo GIF (demo.tape)
-- demo schema    →  used by the try-pgview walkthrough guide

-- ═══════════════════════════════════════════════════════════════
-- public schema  (demo tape: routes, orders)
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE public.routes (
    id         bigserial PRIMARY KEY,
    name       varchar(120) NOT NULL,
    status     text NOT NULL DEFAULT 'active'
                   CHECK (status = ANY (ARRAY['active','inactive','archived'])),
    created_at timestamptz NOT NULL DEFAULT now(),
    tags       text[]
);

CREATE INDEX idx_routes_status ON public.routes USING btree (status);
CREATE INDEX idx_routes_tags   ON public.routes USING gin  (tags);

CREATE TABLE public.orders (
    id        bigserial PRIMARY KEY,
    route_id  bigint REFERENCES public.routes(id) ON DELETE CASCADE,
    amount    numeric(10,2) NOT NULL,
    status    text NOT NULL DEFAULT 'pending'
                  CHECK (status = ANY (ARRAY['pending','processing','failed','complete'])),
    notes     text,
    placed_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_orders_route  ON public.orders(route_id);
CREATE INDEX idx_orders_status ON public.orders(status);

CREATE TABLE public.order_items (
    id       bigserial PRIMARY KEY,
    order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE,
    sku      text NOT NULL,
    qty      int  NOT NULL DEFAULT 1
);

CREATE TABLE public.products (
    id    bigserial PRIMARY KEY,
    name  text NOT NULL,
    price numeric(8,2)
);

CREATE SCHEMA reporting;
CREATE VIEW reporting.daily_order_summary AS
    SELECT placed_at::date AS day,
           COUNT(*)        AS order_count,
           SUM(amount)     AS total
    FROM   public.orders
    GROUP  BY 1;

INSERT INTO public.routes (name, status, tags) VALUES
    ('Alice Johnson', 'active',   '{platform,growth}'),
    ('Bob Smith',     'inactive', '{api}'),
    ('Carol White',   'active',   '{platform,api}'),
    ('Dave Brown',    'archived', '{growth}'),
    ('Eve Martinez',  'active',   '{growth}'),
    ('Frank Lee',     'active',   '{platform}'),
    ('Grace Kim',     'inactive', '{api,growth}');

INSERT INTO public.orders (route_id, amount, status, notes) VALUES
    (1, 120.00, 'complete',    'Delivered on time'),
    (1,  85.50, 'failed',      NULL),
    (3, 200.00, 'processing',  'Awaiting customs'),
    (5,  45.99, 'failed',      NULL),
    (6, 310.00, 'complete',    ''),
    (2,  73.20, 'pending',     NULL),
    (4,  15.00, 'failed',      'Address not found');

INSERT INTO public.order_items (order_id, sku, qty) VALUES
    (1, 'SKU-001', 2), (1, 'SKU-002', 1),
    (2, 'SKU-003', 3), (3, 'SKU-001', 1), (4, 'SKU-004', 5);

INSERT INTO public.products (name, price) VALUES
    ('Widget A', 9.99), ('Widget B', 19.99),
    ('Gadget X', 49.99), ('Gadget Y', 29.99), ('Doohickey', 5.99);

-- ═══════════════════════════════════════════════════════════════
-- demo schema  (try-pgview hands-on walkthrough)
-- ═══════════════════════════════════════════════════════════════

CREATE SCHEMA demo;

CREATE TABLE demo.customers (
    id         bigserial PRIMARY KEY,
    name       text        NOT NULL,
    email      text        NOT NULL UNIQUE,
    status     text        NOT NULL DEFAULT 'active',
    created_at timestamptz NOT NULL DEFAULT now(),
    tags       text[],
    CONSTRAINT customers_status_chk CHECK (status IN ('active','inactive','suspended'))
);
COMMENT ON COLUMN demo.customers.tags IS 'freeform labels for segmentation';
CREATE INDEX idx_customers_status ON demo.customers (status);

CREATE TABLE demo.products (
    id         serial PRIMARY KEY,
    name       text          NOT NULL,
    category   text          NOT NULL,
    price      numeric(10,2) NOT NULL,
    stock_qty  int           NOT NULL DEFAULT 0
);
CREATE INDEX idx_products_category ON demo.products (category);

CREATE TABLE demo.orders (
    id          bigserial PRIMARY KEY,
    customer_id bigint        NOT NULL REFERENCES demo.customers(id),
    total       numeric(10,2) NOT NULL,
    status      text          NOT NULL DEFAULT 'pending',
    created_at  timestamptz   NOT NULL DEFAULT now(),
    notes       jsonb,
    CONSTRAINT orders_status_chk CHECK (status IN ('pending','paid','shipped','cancelled'))
);
CREATE INDEX idx_demo_orders_customer ON demo.orders (customer_id);
CREATE INDEX idx_demo_orders_status   ON demo.orders (status);

CREATE TABLE demo.order_items (
    id         bigserial PRIMARY KEY,
    order_id   bigint        NOT NULL REFERENCES demo.orders(id),
    product_id int           NOT NULL REFERENCES demo.products(id),
    quantity   int           NOT NULL DEFAULT 1,
    unit_price numeric(10,2) NOT NULL
);

INSERT INTO demo.customers (name, email, status, tags) VALUES
    ('Alice Nguyen',    'alice@example.com',   'active',    '{vip,newsletter}'),
    ('Bob Okonkwo',     'bob@example.com',     'active',    '{newsletter}'),
    ('Carol Singh',     'carol@example.com',   'active',    '{vip}'),
    ('David Rossi',     'david@example.com',   'inactive',  NULL),
    ('Eve Martínez',    'eve@example.com',     'active',    '{vip,newsletter}'),
    ('Frank Dubois',    'frank@example.com',   'suspended', NULL),
    ('Grace Kim',       'grace@example.com',   'active',    '{newsletter}'),
    ('Hiro Tanaka',     'hiro@example.com',    'active',    '{vip}'),
    ('Ingrid Svensson', 'ingrid@example.com',  'active',    '{newsletter}'),
    ('James O''Brien',  'james@example.com',   'inactive',  NULL);

INSERT INTO demo.products (name, category, price, stock_qty) VALUES
    ('Wireless Keyboard',   'Electronics', 79.99,  50),
    ('USB-C Hub',           'Electronics', 49.99, 120),
    ('Mechanical Keyboard', 'Electronics', 149.00, 30),
    ('Desk Lamp',           'Office',       35.50, 200),
    ('Notebook A5',         'Stationery',    8.99, 500),
    ('Ergonomic Mouse',     'Electronics',  59.00,  75),
    ('Monitor Stand',       'Office',       89.00,  40),
    ('Cable Organiser',     'Office',       14.99, 300);

INSERT INTO demo.orders (customer_id, total, status, notes) VALUES
    (1, 129.98, 'paid',      '{"source":"web","promo":"SUMMER10"}'),
    (2,  49.99, 'shipped',   NULL),
    (3, 238.00, 'paid',      '{"source":"app"}'),
    (1,  35.50, 'pending',   NULL),
    (5,  59.00, 'paid',      '{"source":"web"}'),
    (7,   8.99, 'cancelled', NULL),
    (8, 168.99, 'shipped',   '{"source":"app","gift":true}'),
    (9,  14.99, 'paid',      NULL);

INSERT INTO demo.order_items (order_id, product_id, quantity, unit_price) VALUES
    (1, 1, 1, 79.99), (1, 2, 1, 49.99),
    (2, 2, 1, 49.99),
    (3, 3, 1, 149.00), (3, 7, 1, 89.00),
    (4, 4, 1, 35.50),
    (5, 6, 1, 59.00),
    (6, 5, 1, 8.99),
    (7, 3, 1, 149.00), (7, 6, 1, 19.99),
    (8, 8, 1, 14.99);

ANALYZE;
