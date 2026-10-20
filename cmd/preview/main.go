package main

import (
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/ui"
)

func main() {
	cols := []Colonist.Colonist{
		Colonist.NewColonist(0, "Alan"),
		Colonist.NewColonist(1, "Timur"),
		Colonist.NewColonist(2, "Sacha"),
	}

	// Пример показателей для проверки полосок.
	cols[0].Hunger = 20
	cols[0].Fatigue = 40

	c := colony.NewColony(cols)

	if err := ui.Run(c, nil); err != nil {
		panic(err)
	}
}
