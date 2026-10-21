package colony

func (c *Colony) aliveCountLocked() int {
	count := 0

	for _, col := range c.colonists {
		if !col.Dead {
			count++
		}
	}
	return count
}
