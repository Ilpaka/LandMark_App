-- +goose Up
SET search_path TO places, public;
INSERT INTO places_categories (slug, title, icon, color, sort_order) VALUES
    ('museum','Музей','museum','#0E7C7B',10),
    ('gastro','Гастро','utensils','#FF6B47',20),
    ('nature','Природа','tree','#7BC9C8',30),
    ('architecture','Архитектура','building','#0E5C5B',40),
    ('viewpoint','Смотровая','eye','#E7A400',50),
    ('park','Парк','leaf','#159594',60);

-- +goose Down
SET search_path TO places, public;
DELETE FROM places_categories WHERE slug IN ('museum','gastro','nature','architecture','viewpoint','park');
