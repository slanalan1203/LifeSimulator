-- +goose Up
CREATE TABLE colonists (
                           village_id BIGINT NOT NULL REFERENCES villages(id),
                           id BIGINT NOT NULL CHECK (id >= 0),
                           name TEXT NOT NULL,
                           age INTEGER NOT NULL CHECK (age >= 0),
                           dead BOOLEAN NOT NULL,
                           hunger SMALLINT NOT NULL CHECK (hunger BETWEEN 0 AND 100),
                           fatigue SMALLINT NOT NULL CHECK (fatigue BETWEEN 0 AND 100),
                           preferred_work TEXT NOT NULL CHECK (preferred_work IN ('food', 'wood')),
                           woodcutting_xp INTEGER NOT NULL CHECK (woodcutting_xp >= 0),
                           gathering_xp INTEGER NOT NULL CHECK (gathering_xp >= 0),
                           food_gathered BIGINT NOT NULL CHECK (food_gathered >= 0),
                           wood_gathered BIGINT NOT NULL CHECK (wood_gathered >= 0),
                           PRIMARY KEY (village_id, id)
);

-- +goose Down
DROP TABLE colonists;