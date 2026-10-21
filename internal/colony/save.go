package colony

import (
	"encoding/json"
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/resources"
	"os"
	"time"
)

type saveData struct {
	Version             int                 `json:"version"`
	Colonist            []Colonist.Colonist `json:"colonist"`
	Food                int                 `json:"food"`
	Wood                int                 `json:"wood"`
	Houses              int                 `json:"houses"`
	Weather             Weather             `json:"weather"`
	ConstructionAllowed bool                `json:"construction_allowed"`
	FoodProduced        int                 `json:"food_produced"`
	FoodConsumed        int                 `json:"food_consumed"`
}

func (c *Colony) Save(path string) error {
	snapshot := c.Snapshot()

	data := &saveData{
		Version:             3,
		Colonist:            snapshot.Colonists,
		Food:                snapshot.Food,
		Wood:                snapshot.Wood,
		Houses:              snapshot.Houses,
		Weather:             snapshot.Weather,
		ConstructionAllowed: snapshot.ConstructionAllowed,
		FoodProduced:        snapshot.FoodProduced,
		FoodConsumed:        snapshot.FoodConsumed,
	}

	content, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return fmt.Errorf("подготовка сохранения: %w", err)
	}

	tempPath := path + ".tmp"
	defer os.Remove(tempPath)
	if err := os.WriteFile(tempPath, content, 0600); err != nil {
		return fmt.Errorf("Запись сохранения: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("Заменить сохранение: %w", err)
	}
	return nil
}

func Load(path string) (*Colony, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("прочитать сохранение: %w", err)
	}
	var data saveData
	if err := json.Unmarshal(content, &data); err != nil {
		return nil, fmt.Errorf("разбираем сохранение: %w", err)
	}

	switch data.Version {
	case 1:
		for i := range data.Colonist {
			data.Colonist[i].Age = 18
		}
	case 2, 3:
	default:
		return nil, fmt.Errorf("неподдерживаемая версия: %d", data.Version)
	}
	if len(data.Colonist) == 0 {
		return nil, fmt.Errorf("В сохранении нет жителей")
	}

	if data.Food < 0 || data.Wood < 0 || data.Houses < 0 || data.FoodProduced < 0 || data.FoodConsumed < 0 {
		return nil, fmt.Errorf("В сохранении есть отрицательные значения")
	}

	if data.Weather != WeatherRain && data.Weather != WeatherClear {
		return nil, fmt.Errorf("неизвестная погода: %q", data.Weather)
	}

	for i := range data.Colonist {
		col := &data.Colonist[i]
		if col.Age < 0 {
			return nil, fmt.Errorf("отрицательный возраст жителя: %d", col.ID)
		}

		col.Busy = false
		col.EmployType = resources.NotSet
		switch {
		case col.Dead:
			col.Action = "умер"
		case col.Age < 18:
			col.Action = "играет"
		default:
			col.Action = "свободен"
		}
		col.ActionStartedAt = time.Time{}
		col.ActionDuration = 0
	}
	c := NewColony(data.Colonist)

	c.storage[resources.Food] = data.Food
	c.storage[resources.Wood] = data.Wood
	c.houses = data.Houses
	c.weather = data.Weather
	c.allowConstruction = data.ConstructionAllowed
	c.foodProduced = data.FoodProduced
	c.foodConsumed = data.FoodConsumed
	return c, nil
}
