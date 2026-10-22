package colony

import "time"

func (c *Colony) NeedsLoop(stop <-chan struct{}) {
	interval := time.Duration(c.Config().Needs.TickSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-c.needsChanged:
			updatedInterval := time.Duration(c.Config().Needs.TickSeconds) * time.Second
			if updatedInterval != interval {
				interval = updatedInterval
				ticker.Reset(interval)
			}
		case <-ticker.C:
			c.mu.Lock()
			count := len(c.colonists)
			for i := range c.colonists {
				col := &c.colonists[i]
				if col.Dead {
					continue
				}
				col.Hunger = min(100, col.Hunger+col.HungerGrowth(c.cfg))
			}
			c.mu.Unlock()

			for i := range count {
				c.Eat(i)
			}
		}
	}
}
