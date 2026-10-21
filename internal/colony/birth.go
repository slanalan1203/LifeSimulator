package colony

import (
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/resources"
	"math/rand/v2"
)

const birthFoodCost = 3

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
	if c.storage[resources.Food] < (population+1)*4+birthFoodCost {
		return false
	}

	adults := 0
	for _, col := range c.colonists {
		if !col.Dead && col.Age >= 18 && col.Age <= 45 {
			adults++
		}
	}

	return adults >= 2
}

func (c *Colony) TryBirth(name string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.canHaveChildLocked() {
		return -1, false
	}

	index := len(c.colonists)
	child := Colonist.NewColonist(index, name)
	child.Age = 0
	child.Action = "играет"
	if rand.IntN(2) == 0 {
		child.PreferredWork = resources.Food
	}

	c.storage[resources.Food] -= birthFoodCost
	c.recordFoodChangedLocked(0, birthFoodCost)
	c.colonists = append(c.colonists, child)
	c.logf("В деревне родился ребёнок: %s", name)

	return index, true
}
