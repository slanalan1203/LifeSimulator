package colony

import "math"

func annualDeathChance(age int) float64 {
	if age == 0 {
		return 0.01
	}
	if age < 15 {
		return 0.0002
	}

	rate := 0.0005 + 0.00009*math.Exp(0.09*float64(age))
	return -math.Expm1(-rate)
}
