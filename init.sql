CREATE DATABASE rinha;

\c rinha;

CREATE TABLE IF NOT EXISTS person(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    nickname VARCHAR(32) NOT NULL UNIQUE,
    birth_date DATE NOT NULL,
    stack TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS person_nickname_idx ON person(nickname); 
CREATE INDEX IF NOT EXISTS person_name_idx ON person(name);
CREATE INDEX IF NOT EXISTS person_stack_idx ON person(stack);