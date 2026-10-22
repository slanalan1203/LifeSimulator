package colony

import "lifeSimulator/internal/balance"

func (c *Colony) Config() balance.Config {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.cfg
}

func (c *Colony) UpdateConfig(cfg balance.Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}

	c.mu.Lock()
	previous := c.cfg
	c.cfg = cfg
	c.mu.Unlock()

	if previous == cfg {
		return nil
	}

	if previous.Needs.TickSeconds != cfg.Needs.TickSeconds {
		select {
		case c.needsChanged <- struct{}{}:
		default:
		}
	}

	if previous.Time.YearSeconds != cfg.Time.YearSeconds {
		select {
		case c.yearChanged <- struct{}{}:
		default:
		}
	}

	c.logf("Настройки деревни изменены")
	return nil
}
