CREATE TABLE IF NOT EXISTS nodes (
    id              UUID PRIMARY KEY,
    parent_id       UUID REFERENCES nodes(id) ON DELETE CASCADE,
    kind            TEXT NOT NULL CHECK (kind IN ('folder', 'file')),
    name            TEXT NOT NULL CHECK (length(name) BETWEEN 1 AND 200),
    description     TEXT NOT NULL DEFAULT '',
    mechanics       TEXT NOT NULL DEFAULT '',
    characteristics JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_nodes_parent_id ON nodes(parent_id);

CREATE TABLE IF NOT EXISTS node_references (
    id            UUID PRIMARY KEY,
    node_id       UUID NOT NULL REFERENCES nodes(id) ON DELETE CASCADE,
    original_name TEXT NOT NULL,
    stored_name   TEXT NOT NULL UNIQUE,
    content_type  TEXT NOT NULL,
    size_bytes    BIGINT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_node_references_node_id ON node_references(node_id);
