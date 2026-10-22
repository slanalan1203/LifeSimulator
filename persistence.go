package main

import (
	"context"
	"errors"
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/database"
	"lifeSimulator/internal/resources"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const villageID int64 = 1
const autosaveInterval = 30 * time.Second
const saveTimeout = 10 * time.Second

func restoreVillage(ctx context.Context, pool *pgxpool.Pool, cfg balance.Config) (*colony.Colony, error) {
	c, err := database.LoadVillage(ctx, pool, villageID, cfg)
	if err == nil {
		c.AddEvent("Деревня загружена из базы данных")
		return c, nil
	}
	if !errors.Is(err, database.ErrVillageNotFound) {
		return nil, err
	}

	c, err = colony.Load(savePath, cfg)
	imported := err == nil
	if errors.Is(err, os.ErrNotExist) {
		colonist1 := Colonist.NewColonist(0, "Alan", cfg.Population.AdultAge)
		colonist2 := Colonist.NewColonist(1, "Timur", cfg.Population.AdultAge)
		colonist2.PreferredWork = resources.Food
		colonist3 := Colonist.NewColonist(2, "Sasha", cfg.Population.AdultAge)
		c = colony.NewColony([]Colonist.Colonist{colonist1, colonist2, colonist3}, cfg)
	} else if err != nil {
		return nil, fmt.Errorf("перенести сохранение в базу данных: %w", err)
	}

	if err := database.SaveVillage(ctx, pool, villageID, c.PersistenceSnapshot()); err != nil {
		return nil, fmt.Errorf("создать сохранение в базе данных: %w", err)
	}
	if imported {
		c.AddEvent("Сохранение деревни перенесено из JSON в базу данных")
	} else {
		c.AddEvent("Новая деревня сохранена в базе данных")
	}
	return c, nil
}

func saveProgress(ctx context.Context, pool *pgxpool.Pool, c *colony.Colony) error {
	snapshot := c.PersistenceSnapshot()
	dbErr := database.SaveVillage(ctx, pool, villageID, snapshot)
	backupErr := colony.SaveSnapshot(savePath, snapshot)
	return errors.Join(dbErr, backupErr)
}

func autosave(ctx context.Context, pool *pgxpool.Pool, c *colony.Colony) {
	ticker := time.NewTicker(autosaveInterval)
	defer ticker.Stop()
	failed := false

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			saveCtx, cancel := context.WithTimeout(ctx, saveTimeout)
			err := saveProgress(saveCtx, pool, c)
			cancel()
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				c.AddEvent(fmt.Sprintf("Ошибка автосохранения: %v", err))
				failed = true
			} else if failed {
				c.AddEvent("Автосохранение снова работает")
				failed = false
			}
		}
	}
}
