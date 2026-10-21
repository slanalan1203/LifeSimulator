package colony

import "lifeSimulator/internal/resources"

func (c *Colony) chooseWork(workerIndex int) resources.Resource {
	c.mu.Lock()
	defer c.mu.Unlock()

	food := c.storage[resources.Food]
	wood := c.storage[resources.Wood]
	people := len(c.colonists)
	col := &c.colonists[workerIndex]

	foodLow := max(10, people*4)
	foodHigh := max(12, people*12)

	woodLow := houseWoodCost * 2
	woodHigh := houseWoodCost * 4

	if food <= foodLow {
		c.restockFood = true
	} else if food >= foodHigh {
		c.restockFood = false
	}

	if wood <= woodLow {
		c.restockWood = true
	} else if wood >= woodHigh {
		c.restockWood = false
	}

	resource := resources.NotSet

	switch {
	case food < people:
		resource = resources.Food
	case c.restockFood && col.PreferredWork == resources.Food:
		resource = resources.Food
	case c.restockWood && col.PreferredWork == resources.Wood:
		resource = resources.Wood
	case c.restockFood:
		resource = resources.Food
	case c.restockWood:
		resource = resources.Wood
	}

	switch resource {
	case resources.Food:
		c.logf(
			"%s пополняет запас еды: сейчас %d, цель %d",
			col.Name, food, foodHigh,
		)
	case resources.Wood:
		c.logf(
			"%s пополняет запас дерева: сейчас %d, цель %d",
			col.Name, wood, woodHigh,
		)
	}
	return resource
}
