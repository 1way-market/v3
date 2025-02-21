-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "btree_gin";

-- Create ads table
CREATE TABLE ads (
    id SERIAL PRIMARY KEY,
    title JSONB NOT NULL,
    description JSONB,
    properties JSONB,
    category_ids INTEGER[],
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    price DECIMAL(15,2),
    search_vector tsvector,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_ads_status ON ads(status);
CREATE INDEX idx_ads_category_ids ON ads USING GIN(category_ids);
CREATE INDEX idx_ads_search_vector ON ads USING GIN(search_vector);
CREATE INDEX idx_ads_title ON ads USING GIN(title jsonb_path_ops);
CREATE INDEX idx_ads_properties ON ads USING GIN(properties);
CREATE INDEX idx_ads_price ON ads(price);
CREATE INDEX idx_ads_created_at ON ads(created_at);

-- Create categories closure table
CREATE TABLE category_closure (
    ancestor_id INTEGER NOT NULL,
    descendant_id INTEGER NOT NULL,
    depth INTEGER NOT NULL,
    PRIMARY KEY (ancestor_id, descendant_id)
);

CREATE INDEX idx_category_closure_ancestor ON category_closure(ancestor_id);
CREATE INDEX idx_category_closure_descendant ON category_closure(descendant_id);

-- Create function to update search vector
CREATE OR REPLACE FUNCTION ads_search_vector_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', COALESCE(NEW.title->>'ru', '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.title->>'en', '')), 'A') ||
        setweight(to_tsvector('russian', COALESCE(NEW.description->>'ru', '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.description->>'en', '')), 'B');
    RETURN NEW;
END
$$ LANGUAGE plpgsql;

-- Create trigger for search vector updates
CREATE TRIGGER ads_search_vector_update
    BEFORE INSERT OR UPDATE ON ads
    FOR EACH ROW
    EXECUTE FUNCTION ads_search_vector_trigger();
