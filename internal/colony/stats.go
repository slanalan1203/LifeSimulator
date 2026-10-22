package colony

import "lifeSimulator/internal/resources"

type Stats struct {
	Population int
	Food       int
	Wood       int
	Houses     int
	Births     uint64
	Deaths     uint64
	Immigrants uint64
}

func (c *Colony) Stats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()

	return Stats{
		Population: c.aliveCountLocked(),
		Food:       c.storage[resources.Food],
		Wood:       c.storage[resources.Wood],
		Houses:     c.houses,
		Births:     c.births,
		Deaths:     c.deaths,
		Immigrants: c.immigrants,
	}
}
