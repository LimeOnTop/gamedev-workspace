-- One 3D model (GLB) per file node.
CREATE TABLE IF NOT EXISTS node_models (
    id            UUID PRIMARY KEY,
    node_id       UUID NOT NULL UNIQUE REFERENCES nodes(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    stored_name   TEXT NOT NULL UNIQUE,
    content_type  TEXT NOT NULL,
    size_bytes    BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
