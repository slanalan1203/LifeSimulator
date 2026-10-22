package colony

import (
	"lifeSimulator/internal/Colonist"
	"math"
	"time"
)

func (c *Colony) aliveCountLocked() int {
	count := 0

	for _, col := range c.colonists {
		if !col.Dead {
			count++
		}
	}
	return count
}

func (c *Colony) foodConsumptionPerMinuteFor(col Colonist.Colonist) float64 {
	interval := time.Duration(c.cfg.Needs.TickSeconds) * time.Second
	ticksPerMinute := float64(time.Minute) / float64(interval)

	return float64(col.HungerGrowth(c.cfg)) *
		ticksPerMinute /
		float64(c.cfg.Needs.HungerPerMeal)
}

func (c *Colony) foodConsumptionPerMinuteLocked() float64 {
	total := 0.0
	for _, col := range c.colonists {
		total += c.foodConsumptionPerMinuteFor(col)
	}
	return total
}

func (c *Colony) foodReserveLocked(duration time.Duration) int {
	return int(math.Ceil(c.foodConsumptionPerMinuteLocked() * duration.Minutes()))
}
