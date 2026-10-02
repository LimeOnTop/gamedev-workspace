ALTER TABLE nodes
    DROP CONSTRAINT IF EXISTS nodes_scenario_asset_check,
    DROP CONSTRAINT IF EXISTS nodes_file_type_check,
    DROP COLUMN IF EXISTS file_type;
