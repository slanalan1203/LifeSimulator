package colony

import (
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/resources"
	"math"
	"math/rand/v2"
	"time"
)

func (c *Colony) BirthAttempts() int {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.reproductiveAdultsLocked() / 2
}

func (c *Colony) reproductiveAdultsLocked() int {
	adults := 0
	for _, col := range c.colonists {
		if !col.Dead && col.Age >= c.cfg.Population.AdultAge && col.Age <= c.cfg.Population.FertilityMaxAge {
			adults++
		}
	}
	return adults
}

func (c *Colony) CanHaveChild() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.canHaveChildLocked()
}

func (c *Colony) canHaveChildLocked() bool {
	population := c.aliveCountLocked()
	if population >= c.houses*residentPerHouse {
		return false
	}
	child := Colonist.Colonist{Age: 0}
	childReserve := int(math.Ceil(c.foodConsumptionPerMinuteFor(child) * 2))
	food := c.storage[resources.Food]
	cost := c.cfg.Population.BirthFoodCost
	if food < cost || food-cost < c.foodReserveLocked(2*time.Minute)+childReserve {
		return false
	}

	return c.reproductiveAdultsLocked() >= 2
}

func (c *Colony) TryBirth(name string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.canHaveChildLocked() {
		return -1, false
	}

	index := len(c.colonists)
	child := Colonist.NewColonist(index, name, 0)
	child.Action = "играет"
	if rand.IntN(2) == 0 {
		child.PreferredWork = resources.Food
	}

	c.storage[resources.Food] -= c.cfg.Population.BirthFoodCost
	c.recordFoodChangedLocked(0, c.cfg.Population.BirthFoodCost)
	c.colonists = append(c.colonists, child)
	c.births++
	c.logf("В деревне родился ребёнок: %s", child.Name)

	return index, true
}
