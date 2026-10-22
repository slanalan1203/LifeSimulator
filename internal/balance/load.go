package balance

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"
)

func Load(path string) (Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Config{}, fmt.Errorf("прочитать настройки: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(content, &cfg); err != nil {
		return Config{}, fmt.Errorf("разобрать настройки: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("неверные настройки: %w", err)
	}
	return cfg, nil
}

func (c Config) Validate() error {
	maxTimerSeconds := int64(math.MaxInt64) / int64(time.Second)
	if c.Time.YearSeconds <= 0 {
		return fmt.Errorf("time.year_seconds должен быть больше 0")
	}
	if int64(c.Time.YearSeconds) > maxTimerSeconds {
		return fmt.Errorf("time.year_seconds слишком большой")
	}

	if c.Population.AdultAge <= 0 {
		return fmt.Errorf("population.adult_age должен быть больше 0")
	}

	if c.Population.FertilityMaxAge < c.Population.AdultAge {
		return fmt.Errorf("fertility_max_age не может быть меньше adult_age")
	}

	chance := c.Population.BirthChancePerPairPerYear
	if math.IsNaN(chance) || math.IsInf(chance, 0) || chance < 0 || chance > 1 {
		return fmt.Errorf("birth_chance_per_pair_per_year должен быть от 0 до 1")
	}

	if c.Population.BirthFoodCost < 0 {
		return fmt.Errorf("birth_food_cost не может быть отрицательным")
	}

	if c.Needs.TickSeconds <= 0 {
		return fmt.Errorf("needs.tick_seconds должен быть больше 0")
	}
	if int64(c.Needs.TickSeconds) > maxTimerSeconds {
		return fmt.Errorf("needs.tick_seconds слишком большой")
	}

	if c.Needs.HungerPerMeal <= 0 || c.Needs.HungerPerMeal > 100 {
		return fmt.Errorf("needs.hunger_per_meal должен быть от 1 до 100")
	}

	if c.Needs.AdultHungerPerTick < 0 || c.Needs.AdultHungerPerTick > 100 {
		return fmt.Errorf("needs.adult_hunger_per_tick должен быть от 0 до 100")
	}

	if c.Needs.ChildHungerPerTick < 0 || c.Needs.ChildHungerPerTick > 100 {
		return fmt.Errorf("needs.child_hunger_per_tick должен быть от 0 до 100")
	}

	growth := max(c.Needs.AdultHungerPerTick, c.Needs.ChildHungerPerTick)
	if growth > c.Needs.HungerPerMeal || c.Needs.HungerPerMeal-1+growth > 100 {
		return fmt.Errorf("прирост голода слишком большой для выбранного размера порции")
	}

	if c.Work.WoodSeconds <= 0 || c.Work.FoodSeconds <= 0 {
		return fmt.Errorf("длительность добычи должна быть больше 0")
	}
	if int64(c.Work.WoodSeconds) > maxTimerSeconds/400 || int64(c.Work.FoodSeconds) > maxTimerSeconds/400 {
		return fmt.Errorf("длительность добычи слишком большая")
	}

	return nil
}
