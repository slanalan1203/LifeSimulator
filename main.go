package main

import (
	"context"
	"errors"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/api"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/namegen"
	"lifeSimulator/internal/resources"
	"lifeSimulator/internal/ui"
	"math/rand/v2"
	"os"
	"sync"
	"time"
)

const savePath = "village.json"

func main() {
	c, err := colony.Load(savePath)

	if errors.Is(err, os.ErrNotExist) {
		colonist1 := Colonist.NewColonist(0, "Alan")
		colonist2 := Colonist.NewColonist(1, "Timur")
		colonist2.PreferredWork = resources.Food
		colonist3 := Colonist.NewColonist(2, "Sasha")
		cols := []Colonist.Colonist{colonist1, colonist2, colonist3}
		c = colony.NewColony(cols)
	} else if err != nil {
		panic(err)
	}
	jobs := make(chan colony.Job, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stop := ctx.Done()
	server, serverDone := api.Start(api.NewHandler(c, jobs), "127.0.0.1:8080")

	var wg sync.WaitGroup
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
			if ctx.Err() != nil || !c.CanHaveChild() || rand.Float64() >= 0.25 {
				return
			}
			name, err := namegen.Random(ctx)
			if err != nil || ctx.Err() != nil {
				return
			}
			index, ok := c.TryBirth(name)
			if !ok {
				return
			}
			wg.Go(func() {
				c.Live(ctx, index, jobs)
			})
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

				name, err := namegen.Random(ctx)
				if ctx.Err() != nil {
					return
				}
				if err != nil {
					continue
				}
				index, ok := c.TryAddColonist(name)
				if !ok {
					continue
				}

				wg.Go(func() {
					c.Live(ctx, index, jobs)
				})
			}
		}
	})

	uiErr := ui.Run(c, jobs)

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 3*time.Second)
	shutdownErr := server.Shutdown(shutdownCtx)
	shutdownCancel()

	if shutdownErr != nil {
		shutdownErr = errors.Join(shutdownErr, server.Close())
	}
	serveErr := <-serverDone

	wg.Wait()

	saveErr := c.Save(savePath)
	if err := errors.Join(uiErr, shutdownErr, serveErr, saveErr); err != nil {
		panic(err)
	}
}
