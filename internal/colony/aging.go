package colony

import (
	"lifeSimulator/internal/resources"
	"math/rand/v2"
	"time"
)

const yearDuration = 30 * time.Second

func (c *Colony) AgingLoop(stop <-chan struct{}, onYear func()) {
	ticker := time.NewTicker(yearDuration)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
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
					col.Busy = false
					col.EmployType = resources.NotSet
					col.Action = "умер"
					col.ActionStartedAt = time.Time{}
					col.ActionDuration = 0
					c.logf("%s умер, возраст: %d", col.Name, col.Age)
					continue
				}
				if col.Age == 18 {
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
