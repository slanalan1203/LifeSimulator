package balance

type Config struct {
	Time       TimeConfig       `json:"time"`
	Population PopulationConfig `json:"population"`
	Needs      NeedsConfig      `json:"needs"`
	Work       WorkConfig       `json:"work"`
}

type TimeConfig struct {
	YearSeconds int `json:"year_seconds"`
}

type PopulationConfig struct {
	AdultAge                  int     `json:"adult_age"`
	FertilityMaxAge           int     `json:"fertility_max_age"`
	BirthChancePerPairPerYear float64 `json:"birth_chance_per_pair_per_year"`
	BirthFoodCost             int     `json:"birth_food_cost"`
}

type NeedsConfig struct {
	TickSeconds        int `json:"tick_seconds"`
	AdultHungerPerTick int `json:"adult_hunger_per_tick"`
	ChildHungerPerTick int `json:"child_hunger_per_tick"`
	HungerPerMeal      int `json:"hunger_per_meal"`
}

type WorkConfig struct {
	WoodSeconds int `json:"wood_seconds"`
	FoodSeconds int `json:"food_seconds"`
}
