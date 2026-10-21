package colony

import "time"

func (c *Colony) NeedsLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			c.mu.Lock()
			count := len(c.colonists)
			for i := range c.colonists {
				col := &c.colonists[i]
				if col.Dead {
					continue
				}
				col.Hunger = min(100, col.Hunger+10)
			}
			c.mu.Unlock()

			for i := range count {
				c.Eat(i)
			}
		}
	}
}
