package colony

import (
	"encoding/json"
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/balance"
	"os"
	"path/filepath"
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
	return SaveSnapshot(path, c.PersistenceSnapshot())
}

func SaveSnapshot(path string, snapshot Snapshot) error {
	if err := ValidateSnapshot(snapshot); err != nil {
		return fmt.Errorf("проверить сохранение: %w", err)
	}

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

	temp, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*.tmp")
	if err != nil {
		return fmt.Errorf("создать временное сохранение: %w", err)
	}
	tempPath := temp.Name()
	defer temp.Close()
	defer os.Remove(tempPath)
	if _, err := temp.Write(content); err != nil {
		return fmt.Errorf("записать сохранение: %w", err)
	}
	if err := temp.Sync(); err != nil {
		return fmt.Errorf("синхронизировать сохранение: %w", err)
	}
	if err := temp.Close(); err != nil {
		return fmt.Errorf("закрыть временное сохранение: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("заменить сохранение: %w", err)
	}
	return nil
}

func Load(path string, cfg balance.Config) (*Colony, error) {
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
			data.Colonist[i].Age = cfg.Population.AdultAge
		}
	case 2, 3:
	default:
		return nil, fmt.Errorf("неподдерживаемая версия: %d", data.Version)
	}
	return FromSnapshot(Snapshot{
		Colonists:           data.Colonist,
		Food:                data.Food,
		Wood:                data.Wood,
		Houses:              data.Houses,
		Weather:             data.Weather,
		ConstructionAllowed: data.ConstructionAllowed,
		FoodProduced:        data.FoodProduced,
		FoodConsumed:        data.FoodConsumed,
	}, cfg)
}
