CREATE TABLE IF NOT EXISTS asset_categories (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order  INT  NOT NULL DEFAULT 0
);

INSERT INTO asset_categories (id, name, description, sort_order) VALUES
    ('weapon',    'Оружие',                    'Ручное оружие ближнего и дальнего боя, метательное оружие, щиты.', 10),
    ('siege',     'Осадные орудия',            'Катапульты, требушеты, тараны, баллисты, пушки и другие крупные машины.', 20),
    ('map',       'Карта',                     'Локации, постройки, помещения и крупные части уровня.', 30),
    ('character', 'Персонажи',                 'Игровые персонажи и NPC.', 40),
    ('item',      'Предметы и припасы',        'Подбираемые и переносимые предметы: сундуки, бочки, ящики, припасы.', 50),
    ('prop',      'Пропсы окружения',          'Мебель, декор и разрушаемые элементы окружения.', 60),
    ('gear',      'Экипировка и кастомизация', 'Броня, шлемы, одежда и косметические предметы.', 70)
ON CONFLICT (id) DO NOTHING;

ALTER TABLE nodes
    ADD COLUMN IF NOT EXISTS asset_category TEXT REFERENCES asset_categories(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS reference_prompt TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_nodes_asset_category ON nodes(asset_category) WHERE asset_category IS NOT NULL;
