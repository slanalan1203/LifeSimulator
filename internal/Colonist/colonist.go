package Colonist

import (
	"lifeSimulator/internal/resources"
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

func NewColonist(id int, name string) Colonist {
	return Colonist{
		ID:            id,
		Name:          name,
		EmployType:    resources.NotSet,
		Busy:          false,
		Age:           18,
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
