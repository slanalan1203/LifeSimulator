package main

import (
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/ui"
)

const configPath = "internal/balance/config.json"

func main() {
	cfg, err := balance.Load(configPath)
	if err != nil {
		panic(err)
	}
	cols := []Colonist.Colonist{
		Colonist.NewColonist(0, "Alan", cfg.Population.AdultAge),
		Colonist.NewColonist(1, "Timur", cfg.Population.AdultAge),
		Colonist.NewColonist(2, "Sacha", cfg.Population.AdultAge),
	}

	cols[0].Hunger = 20
	cols[0].Fatigue = 40

	c := colony.NewColony(cols, cfg)

	if err := ui.Run(c, nil, configPath); err != nil {
		panic(err)
	}
}
