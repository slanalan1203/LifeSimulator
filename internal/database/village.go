package database

import (
	"context"
	"errors"
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/resources"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrVillageNotFound = errors.New("деревня не найдена")

func SaveVillage(
	ctx context.Context,
	pool *pgxpool.Pool,
	villageID int64,
	snapshot colony.Snapshot,
) error {
	if err := colony.ValidateSnapshot(snapshot); err != nil {
		return fmt.Errorf("проверить снимок деревни: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("начать сохранение: %w", err)
	}
	defer rollbackVillageTx(tx)

	_, err = tx.Exec(ctx, `
		INSERT INTO villages (
			id, food, wood, houses, weather,
			construction_allowed, food_produced, food_consumed
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			food = EXCLUDED.food,
			wood = EXCLUDED.wood,
			houses = EXCLUDED.houses,
			weather = EXCLUDED.weather,
			construction_allowed = EXCLUDED.construction_allowed,
			food_produced = EXCLUDED.food_produced,
			food_consumed = EXCLUDED.food_consumed,
			updated_at = now()
	`,
		villageID,
		snapshot.Food,
		snapshot.Wood,
		snapshot.Houses,
		string(snapshot.Weather),
		snapshot.ConstructionAllowed,
		snapshot.FoodProduced,
		snapshot.FoodConsumed,
	)
	if err != nil {
		return fmt.Errorf("сохранить деревню: %w", err)
	}

	_, err = tx.Exec(ctx,
		"DELETE FROM colonists WHERE village_id = $1",
		villageID,
	)
	if err != nil {
		return fmt.Errorf("заменить прежний список жителей: %w", err)
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"colonists"},
		[]string{
			"village_id", "id", "name", "age", "dead",
			"hunger", "fatigue", "preferred_work",
			"woodcutting_xp", "gathering_xp",
			"food_gathered", "wood_gathered",
		},
		pgx.CopyFromSlice(len(snapshot.Colonists), func(i int) ([]any, error) {
			col := snapshot.Colonists[i]

			return []any{
				villageID, col.ID, col.Name, col.Age, col.Dead,
				col.Hunger, col.Fatigue, string(col.PreferredWork),
				col.Woodcutting.XP, col.Gathering.XP,
				col.FoodGathered, col.WoodGathered,
			}, nil
		}),
	)
	if err != nil {
		return fmt.Errorf("сохранить жителей: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("завершить сохранение: %w", err)
	}

	return nil
}

func LoadVillage(
	ctx context.Context,
	pool *pgxpool.Pool,
	villageID int64,
	cfg balance.Config,
) (*colony.Colony, error) {
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.RepeatableRead,
		AccessMode: pgx.ReadOnly,
	})
	if err != nil {
		return nil, fmt.Errorf("начать загрузку деревни: %w", err)
	}
	defer rollbackVillageTx(tx)

	var snapshot colony.Snapshot
	var weather string
	err = tx.QueryRow(ctx, `
		SELECT food, wood, houses, weather, construction_allowed,
			food_produced, food_consumed
		FROM villages
		WHERE id = $1
	`, villageID).Scan(
		&snapshot.Food,
		&snapshot.Wood,
		&snapshot.Houses,
		&weather,
		&snapshot.ConstructionAllowed,
		&snapshot.FoodProduced,
		&snapshot.FoodConsumed,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("загрузить деревню %d: %w", villageID, ErrVillageNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("загрузить деревню: %w", err)
	}
	snapshot.Weather = colony.Weather(weather)

	rows, err := tx.Query(ctx, `
		SELECT id, name, age, dead, hunger, fatigue, preferred_work,
			woodcutting_xp, gathering_xp, food_gathered, wood_gathered
		FROM colonists
		WHERE village_id = $1
		ORDER BY id
	`, villageID)
	if err != nil {
		return nil, fmt.Errorf("загрузить жителей: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col Colonist.Colonist
		var preferredWork string
		if err := rows.Scan(
			&col.ID,
			&col.Name,
			&col.Age,
			&col.Dead,
			&col.Hunger,
			&col.Fatigue,
			&preferredWork,
			&col.Woodcutting.XP,
			&col.Gathering.XP,
			&col.FoodGathered,
			&col.WoodGathered,
		); err != nil {
			return nil, fmt.Errorf("прочитать жителя: %w", err)
		}
		col.PreferredWork = resources.Resource(preferredWork)
		snapshot.Colonists = append(snapshot.Colonists, col)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("прочитать список жителей: %w", err)
	}
	rows.Close()

	c, err := colony.FromSnapshot(snapshot, cfg)
	if err != nil {
		return nil, fmt.Errorf("восстановить деревню: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("завершить загрузку: %w", err)
	}

	return c, nil
}

func rollbackVillageTx(tx pgx.Tx) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = tx.Rollback(ctx)
}
