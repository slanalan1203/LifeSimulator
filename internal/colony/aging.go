package colony

import (
	"lifeSimulator/internal/resources"
	"math/rand/v2"
	"time"
)

func (c *Colony) AgingLoop(stop <-chan struct{}, onYear func()) {
	interval := time.Duration(c.Config().Time.YearSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-c.yearChanged:
			updatedInterval := time.Duration(c.Config().Time.YearSeconds) * time.Second
			if updatedInterval != interval {
				interval = updatedInterval
				ticker.Reset(interval)
			}
		case <-ticker.C:
			c.mu.Lock()

			for i := range c.colonists {
				col := &c.colonists[i]
				if col.Dead {
					continue
				}
				chance := annualDeathChance(col.Age)
				col.Age++
				if rand.Float64() < chance {
					col.Dead = true
					c.deaths++
					col.Busy = false
					col.EmployType = resources.NotSet
					col.Action = "умер"
					col.ActionStartedAt = time.Time{}
					col.ActionDuration = 0
					c.logf("%s умер, возраст: %d", col.Name, col.Age)
					continue
				}
				if col.Age == c.cfg.Population.AdultAge {
					c.logf("%s стал взрослым и теперь может работать", col.Name)
				} else {
					c.logf("День рождения: %s, возраст %d", col.Name, col.Age)
				}
			}
			c.mu.Unlock()
			if onYear != nil {
				onYear()
			}
		}
	}
}
