package Colonist

import (
	"fmt"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/resources"
	"strings"
	"time"
)

type Colonist struct {
	ID         int
	Name       string
	EmployType resources.Resource
	Busy       bool
	Age        int
	Dead       bool

	Hunger  int
	Fatigue int

	PreferredWork resources.Resource

	Action          string
	ActionStartedAt time.Time
	ActionDuration  time.Duration

	Woodcutting Skill
	Gathering   Skill

	FoodGathered int
	WoodGathered int
}

func NewColonist(id int, name string, age int) Colonist {
	name = strings.TrimSpace(name)
	if name == "" {
		name = fmt.Sprintf("Житель №%d", id)
	}
	return Colonist{
		ID:            id,
		Name:          name,
		EmployType:    resources.NotSet,
		Busy:          false,
		Age:           age,
		PreferredWork: resources.Wood,

		Action: "свободен",
	}
}

func (c *Colonist) SetBusy(b bool) {
	c.Busy = b
}

func (c Colonist) Level() int {
	return max(c.Woodcutting.Level(), c.Gathering.Level())
}

func (c Colonist) HungerGrowth(cfg balance.Config) int {
	if c.Dead {
		return 0
	}
	if c.Age < cfg.Population.AdultAge {
		return cfg.Needs.ChildHungerPerTick
	}
	return cfg.Needs.AdultHungerPerTick
}
