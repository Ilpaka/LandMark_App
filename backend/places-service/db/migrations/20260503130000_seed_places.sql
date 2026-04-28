-- +goose Up
SET search_path TO places, public;

-- Seed публичных POI (из дизайн-мокапов)
INSERT INTO places_places (id, title, description, latitude, longitude, city, country, status, source, published_at, created_at, updated_at)
VALUES
    (gen_random_uuid(), 'Казанский Кремль', 'Историческая крепость в центре Казани, объект ЮНЕСКО', 55.7986, 49.1047, 'Казань', 'Россия', 'published', 'seed', now(), now(), now()),
    (gen_random_uuid(), 'Эрмитаж', 'Один из крупнейших художественных музеев мира', 59.9398, 30.3146, 'Санкт-Петербург', 'Россия', 'published', 'seed', now(), now(), now()),
    (gen_random_uuid(), 'Красная площадь', 'Главная площадь Москвы, символ России', 55.7539, 37.6208, 'Москва', 'Россия', 'published', 'seed', now(), now(), now()),
    (gen_random_uuid(), 'Байкал', 'Самое глубокое озеро в мире, объект ЮНЕСКО', 53.5587, 108.1650, 'Иркутск', 'Россия', 'published', 'seed', now(), now(), now()),
    (gen_random_uuid(), 'Петергоф', 'Дворцово-парковый ансамбль с фонтанами', 59.8845, 29.9025, 'Петергоф', 'Россия', 'published', 'seed', now(), now(), now()),
    (gen_random_uuid(), 'Мечеть Кул Шариф', 'Главная мечеть Татарстана в Казанском Кремле', 55.7993, 49.1004, 'Казань', 'Россия', 'published', 'seed', now(), now(), now());

-- Привязать категории seed-мест
WITH museum_cat AS (SELECT id FROM places_categories WHERE slug='museum'),
     arch_cat AS (SELECT id FROM places_categories WHERE slug='architecture'),
     nature_cat AS (SELECT id FROM places_categories WHERE slug='nature')
INSERT INTO places_place_categories (place_id, category_id)
SELECT p.id, museum_cat.id FROM places_places p, museum_cat WHERE p.title IN ('Эрмитаж', 'Казанский Кремль')
UNION ALL
SELECT p.id, arch_cat.id FROM places_places p, arch_cat WHERE p.title IN ('Красная площадь', 'Петергоф', 'Казанский Кремль', 'Мечеть Кул Шариф')
UNION ALL
SELECT p.id, nature_cat.id FROM places_places p, nature_cat WHERE p.title IN ('Байкал')
ON CONFLICT DO NOTHING;

-- +goose Down
SET search_path TO places, public;
DELETE FROM places_places WHERE source='seed';
