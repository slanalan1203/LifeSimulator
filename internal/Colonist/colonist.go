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

	Hunger  int
	Fatigue int

	PreferredWork resources.Resource

	Action          string
	ActionStartedAt time.Time
	ActionDuration  time.Duration
}

func NewColonist(id int, name string) Colonist {
	return Colonist{
		ID:         id,
		Name:       name,
		EmployType: resources.NotSet,
		Busy:       false,

		PreferredWork: resources.Wood,

		Action: "свободен",
	}
}

func (c *Colonist) SetBusy(b bool) {
	c.Busy = b
}
