-- +goose Up
CREATE TABLE villages (
    id BIGINT PRIMARY KEY,
    food BIGINT NOT NULL CHECK (food >= 0),
    wood BIGINT NOT NULL CHECK (wood >= 0),
    houses INTEGER NOT NULL CHECK (houses >= 0),
    weather TEXT NOT NULL,
    construction_allowed BOOLEAN NOT NULL,
    food_produced BIGINT NOT NULL CHECK (food_produced >= 0),
    food_consumed BIGINT NOT NULL CHECK (food_consumed >= 0),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE villages;
