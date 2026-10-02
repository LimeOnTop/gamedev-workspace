-- Files are either object cards (the default) or scenario documents; folders have no file type.
ALTER TABLE nodes ADD COLUMN IF NOT EXISTS file_type TEXT DEFAULT 'object';

UPDATE nodes SET file_type = NULL WHERE kind = 'folder';
UPDATE nodes SET file_type = 'object' WHERE kind = 'file' AND file_type IS NULL;

ALTER TABLE nodes
    ADD CONSTRAINT nodes_file_type_check CHECK (
        (kind = 'folder' AND file_type IS NULL)
        OR (kind = 'file' AND file_type IN ('object', 'scenario'))
    ),
    -- Scenario files are design documents and never belong to the 3D asset catalog.
    ADD CONSTRAINT nodes_scenario_asset_check CHECK (file_type IS DISTINCT FROM 'scenario' OR asset_category IS NULL);
