CREATE DATABASE rinha;

\c rinha;

CREATE TABLE IF NOT EXISTS person(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    nickname VARCHAR(32) NOT NULL UNIQUE,
    birth_date DATE NOT NULL,
    stack TEXT NOT NULL,
    idx tsvector GENERATED ALWAYS AS (to_tsvector('simple', name || ' ' || nickname || ' ' || stack)) STORED
);

CREATE INDEX idx_person ON person USING GIN(idx);

ALTER SYSTEM SET shared_buffers TO '368MB';

