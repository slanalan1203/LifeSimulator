package colony

import "time"

const yearDuration = 30 * time.Second

func (c *Colony) AgingLoop(stop <-chan struct{}) {
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
				col.Age++
				c.logf("День рождения: %s, возраст %d", col.Name, col.Age)
			}
			c.mu.Unlock()
		}
	}
}
