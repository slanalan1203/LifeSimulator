package main

import (
	"context"
	"errors"
	"lifeSimulator/internal/api"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/database"
	"lifeSimulator/internal/namegen"
	"lifeSimulator/internal/ui"
	"math/rand/v2"
	"os"
	"sync"
	"time"
)

const savePath = "village.json"
const configPath = "internal/balance/config.json"

func main() {

	dbCtx, cancelDB := context.WithTimeout(context.Background(), 5*time.Second)

	pool, err := database.Connect(dbCtx, os.Getenv("DATABASE_URL"))
	cancelDB()

	if err != nil {
		panic(err)
	}
	defer pool.Close()

	cfg, err := balance.Load(configPath)
	if err != nil {
		panic(err)
	}

	loadCtx, cancelLoad := context.WithTimeout(context.Background(), 15*time.Second)
	c, err := restoreVillage(loadCtx, pool, cfg)
	cancelLoad()
	if err != nil {
		panic(err)
	}
	jobs := make(chan colony.Job, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := ctx.Done()
	address := os.Getenv("HTTP_ADDR")
	if address == "" {
		address = "127.0.0.1:8080"
	}
	server, serverDone := api.Start(api.NewHandler(c, jobs), address)

	var wg sync.WaitGroup
	names := namegen.NewPool()
	wg.Go(func() {
		names.Run(ctx)
	})

	initialSnaphot := c.Snapshot()
	for i, col := range initialSnaphot.Colonists {
		if col.Dead {
			continue
		}
		wg.Go(func() {
			c.Live(ctx, i, jobs)
		})
	}

	wg.Go(func() {
		c.WeatherLoop(stop)
	})

	wg.Go(func() {
		c.NeedsLoop(stop)
	})

	wg.Go(func() {
		c.AgingLoop(stop, func() {
			for range c.BirthAttempts() {
				if ctx.Err() != nil || !c.CanHaveChild() {
					return
				}
				if rand.Float64() >= c.Config().Population.BirthChancePerPairPerYear {
					continue
				}
				index, ok := c.TryBirth(names.Next())
				if !ok {
					return
				}
				wg.Go(func() {
					c.Live(ctx, index, jobs)
				})
			}
		})
	})

	wg.Go(func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				snapshot := c.Snapshot()

				if snapshot.Population >= snapshot.Capacity || snapshot.Food < 10 {
					continue
				}
				if rand.IntN(2) != 0 {
					continue
				}

				if ctx.Err() != nil {
					return
				}
				index, ok := c.TryAddColonist(names.Next())
				if !ok {
					continue
				}

				wg.Go(func() {
					c.Live(ctx, index, jobs)
				})
			}
		}
	})

	wg.Go(func() {
		autosave(ctx, pool, c)
	})

	uiErr := ui.Run(c, jobs, configPath)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	shutdownErr := server.Shutdown(shutdownCtx)
	shutdownCancel()

	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, server.Close())
	}
	serveErr := <-serverDone

	wg.Wait()

	saveCtx, cancelSave := context.WithTimeout(context.Background(), saveTimeout)
	saveErr := saveProgress(saveCtx, pool, c)
	cancelSave()
	if err := errors.Join(uiErr, shutdownErr, serveErr, saveErr); err != nil {
		panic(err)
	}
}
