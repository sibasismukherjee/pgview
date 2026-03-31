#!/usr/bin/env bash
set -e

echo "→ stopping any existing container..."
docker rm -f pgview-demo 2>/dev/null || true

echo "→ starting postgres..."
docker run -d --name pgview-demo \
  -e POSTGRES_PASSWORD=demo \
  -e POSTGRES_DB=demodb \
  -p 5433:5432 \
  postgres:16-alpine

echo "→ waiting for demodb to be ready..."
until docker exec pgview-demo psql -U postgres -d demodb -c '\q' 2>/dev/null; do
  sleep 0.5
done

echo "→ creating schema and tables..."
docker exec pgview-demo psql -U postgres -d demodb -c "CREATE SCHEMA reporting;"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE TABLE public.routes (id bigserial PRIMARY KEY, name varchar(120) NOT NULL, status text NOT NULL DEFAULT 'active' CHECK (status = ANY (ARRAY['active','inactive','archived'])), created_at timestamptz NOT NULL DEFAULT now(), tags text[]);"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE INDEX idx_routes_status ON public.routes USING btree (status);"
docker exec pgview-demo psql -U postgres -d demodb -c "CREATE INDEX idx_routes_tags ON public.routes USING gin (tags);"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE TABLE public.orders (id bigserial PRIMARY KEY, route_id bigint REFERENCES public.routes(id) ON DELETE CASCADE, amount numeric(10,2) NOT NULL, placed_at timestamptz NOT NULL DEFAULT now());"
docker exec pgview-demo psql -U postgres -d demodb -c "CREATE INDEX idx_orders_route ON public.orders(route_id);"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE TABLE public.order_items (id bigserial PRIMARY KEY, order_id bigint REFERENCES public.orders(id) ON DELETE CASCADE, sku text NOT NULL, qty int NOT NULL DEFAULT 1);"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE TABLE public.products (id bigserial PRIMARY KEY, name text NOT NULL, price numeric(8,2));"

docker exec pgview-demo psql -U postgres -d demodb -c "CREATE VIEW reporting.daily_order_summary AS SELECT placed_at::date AS day, COUNT(*) AS order_count, SUM(amount) AS total FROM public.orders GROUP BY 1;"

echo "→ inserting data..."
docker exec pgview-demo psql -U postgres -d demodb -c "INSERT INTO public.routes (name, status, tags) VALUES ('Alice Johnson', 'active', '{platform,growth}'), ('Bob Smith', 'inactive', '{api}'), ('Carol White', 'active', '{platform,api}'), ('Dave Brown', 'archived', '{growth}'), ('Eve Martinez', 'active', '{growth}'), ('Frank Lee', 'active', '{platform}'), ('Grace Kim', 'inactive', '{api,growth}');"

docker exec pgview-demo psql -U postgres -d demodb -c "INSERT INTO public.orders (route_id, amount) VALUES (1,120.00),(1,85.50),(3,200.00),(5,45.99),(6,310.00);"

docker exec pgview-demo psql -U postgres -d demodb -c "INSERT INTO public.order_items (order_id, sku, qty) VALUES (1,'SKU-001',2),(1,'SKU-002',1),(2,'SKU-003',3),(3,'SKU-001',1),(4,'SKU-004',5);"

docker exec pgview-demo psql -U postgres -d demodb -c "INSERT INTO public.products (name, price) VALUES ('Widget A',9.99),('Widget B',19.99),('Gadget X',49.99),('Gadget Y',29.99),('Doohickey',5.99);"

echo "→ verifying..."
docker exec pgview-demo psql -U postgres -d demodb -c "\dt public.*"

echo "✓ pgview-demo ready on localhost:5433"
