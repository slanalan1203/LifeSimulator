package colony

import (
	"fmt"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/resources"
	"time"
)

func ValidateSnapshot(snapshot Snapshot) error {
	if len(snapshot.Colonists) == 0 {
		return fmt.Errorf("в сохранении нет жителей")
	}

	if snapshot.Food < 0 || snapshot.Wood < 0 || snapshot.Houses < 0 ||
		snapshot.FoodProduced < 0 || snapshot.FoodConsumed < 0 {
		return fmt.Errorf("в сохранении есть отрицательные ресурсы или счётчики")
	}

	if snapshot.Weather != WeatherRain && snapshot.Weather != WeatherClear {
		return fmt.Errorf("неизвестная погода: %q", snapshot.Weather)
	}

	for i, col := range snapshot.Colonists {
		if col.ID != i {
			return fmt.Errorf("номер жителя %d не соответствует позиции %d", col.ID, i)
		}
		if col.Age < 0 {
			return fmt.Errorf("отрицательный возраст жителя: %d", col.ID)
		}
		if col.Hunger < 0 || col.Hunger > 100 || col.Fatigue < 0 || col.Fatigue > 100 {
			return fmt.Errorf("голод или усталость жителя %d вне диапазона 0–100", col.ID)
		}
		if col.PreferredWork != resources.Food && col.PreferredWork != resources.Wood {
			return fmt.Errorf("неизвестное любимое занятие жителя %d: %q", col.ID, col.PreferredWork)
		}
		if col.Woodcutting.XP < 0 || col.Gathering.XP < 0 ||
			col.FoodGathered < 0 || col.WoodGathered < 0 {
			return fmt.Errorf("отрицательный опыт или добыча жителя: %d", col.ID)
		}
	}

	return nil
}

func FromSnapshot(snapshot Snapshot, cfg balance.Config) (*Colony, error) {
	if err := ValidateSnapshot(snapshot); err != nil {
		return nil, err
	}

	c := NewColony(snapshot.Colonists, cfg)
	for i := range c.colonists {
		col := &c.colonists[i]
		col.Busy = false
		col.EmployType = resources.NotSet
		switch {
		case col.Dead:
			col.Action = "умер"
		case col.Age < cfg.Population.AdultAge:
			col.Action = "играет"
		default:
			col.Action = "свободен"
		}
		col.ActionStartedAt = time.Time{}
		col.ActionDuration = 0
	}

	c.storage[resources.Food] = snapshot.Food
	c.storage[resources.Wood] = snapshot.Wood
	c.houses = snapshot.Houses
	c.weather = snapshot.Weather
	c.allowConstruction = snapshot.ConstructionAllowed
	c.foodProduced = snapshot.FoodProduced
	c.foodConsumed = snapshot.FoodConsumed

	return c, nil
}

func (c *Colony) PersistenceSnapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	snapshot := c.snapshotLocked()
	if c.buildingHouse {
		snapshot.Wood += houseWoodCost
	}

	return snapshot
}
