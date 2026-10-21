package colony

import "time"

type foodChange struct {
	at       time.Time
	produced int
	consumed int
}

func (c *Colony) recordFoodChangedLocked(produced, consumed int) {
	now := time.Now()
	cutoff := now.Add(-time.Minute)
	c.foodProduced += produced
	c.foodConsumed += consumed

	first := 0
	for first < len(c.foodHistory) && !c.foodHistory[first].at.After(cutoff) {
		first++
	}
	c.foodHistory = c.foodHistory[first:]

	c.foodHistory = append(c.foodHistory, foodChange{
		at:       now,
		produced: produced,
		consumed: consumed,
	})
}
