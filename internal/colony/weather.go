package colony

import (
	"math/rand/v2"
	"time"
)

type Weather string

const (
	WeatherClear Weather = "ясно"
	WeatherRain  Weather = "дождь"
)

func (c *Colony) WeatherLoop(stop <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			next := WeatherClear
			if rand.IntN(2) == 0 {
				next = WeatherRain
			}

			c.mu.Lock()
			changed := c.weather != next
			c.weather = next
			c.mu.Unlock()

			if changed {
				c.logf("Погода изменилась: %s", next)
			}
		}
	}
}
