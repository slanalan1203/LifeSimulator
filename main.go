package main

import (
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/namegen"
	"lifeSimulator/internal/resources"
	"lifeSimulator/internal/ui"
	"math/rand/v2"
	"sync"
	"time"
)

func main() {
	colonist1 := Colonist.NewColonist(0, "Alan")
	colonist2 := Colonist.NewColonist(1, "Timur")
	colonist2.PreferredWork = resources.Food
	colonist3 := Colonist.NewColonist(2, "Sacha")
	cols := []Colonist.Colonist{colonist1, colonist2, colonist3}
	c := colony.NewColony(cols)
	jobs := make(chan colony.Job, 10)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := range cols {
		wg.Go(func() {
			c.Live(i, jobs, stop)
		})
	}

	wg.Go(func() {
		c.WeatherLoop(stop)
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

				if len(snapshot.Colonists) >= 5 {
					return
				}

				if !snapshot.HasHouse || snapshot.Food < 10 {
					continue
				}
				if rand.IntN(2) != 0 {
					continue
				}

				name, err := namegen.Random()
				if err != nil {
					continue
				}
				index, ok := c.TryAddColonist(name)
				if !ok {
					continue
				}

				wg.Go(func() {
					c.Live(index, jobs, stop)
				})
			}
		}
	})

	err := ui.Run(c, jobs)
	close(stop)
	wg.Wait()
	close(jobs)

	if err != nil {
		panic(err)
	}
}
