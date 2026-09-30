DROP INDEX IF EXISTS idx_nodes_asset_category;

ALTER TABLE nodes
    DROP COLUMN IF EXISTS reference_prompt,
    DROP COLUMN IF EXISTS asset_category;

DROP TABLE IF EXISTS asset_categories;
