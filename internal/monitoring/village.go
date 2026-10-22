package monitoring

import (
	"lifeSimulator/internal/colony"

	"github.com/prometheus/client_golang/prometheus"
)

type villageCollector struct {
	village    *colony.Colony
	population *prometheus.Desc
	resources  *prometheus.Desc
	houses     *prometheus.Desc
	births     *prometheus.Desc
	deaths     *prometheus.Desc
	immigrants *prometheus.Desc
}

func newVillageCollector(c *colony.Colony) *villageCollector {
	return &villageCollector{
		village: c,
		population: prometheus.NewDesc(
			"village_population",
			"Current number of living residents.",
			nil, nil,
		),
		resources: prometheus.NewDesc(
			"village_resources",
			"Current resource stock.",
			[]string{"resource"}, nil,
		),
		houses: prometheus.NewDesc(
			"village_houses",
			"Current number of houses.",
			nil, nil,
		),
		births: prometheus.NewDesc(
			"village_births_total",
			"Number of births since the simulator started.",
			nil, nil,
		),
		deaths: prometheus.NewDesc(
			"village_deaths_total",
			"Number of resident deaths since the simulator started.",
			nil, nil,
		),
		immigrants: prometheus.NewDesc(
			"village_immigrants_total",
			"Number of residents who immigrated since the simulator started.",
			nil, nil,
		),
	}
}

func (v *villageCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- v.population
	ch <- v.resources
	ch <- v.houses
	ch <- v.births
	ch <- v.deaths
	ch <- v.immigrants
}

func (v *villageCollector) Collect(ch chan<- prometheus.Metric) {
	stats := v.village.Stats()

	ch <- prometheus.MustNewConstMetric(
		v.population, prometheus.GaugeValue,
		float64(stats.Population),
	)
	ch <- prometheus.MustNewConstMetric(
		v.resources, prometheus.GaugeValue,
		float64(stats.Food), "food",
	)
	ch <- prometheus.MustNewConstMetric(
		v.resources, prometheus.GaugeValue,
		float64(stats.Wood), "wood",
	)
	ch <- prometheus.MustNewConstMetric(
		v.houses, prometheus.GaugeValue,
		float64(stats.Houses),
	)
	ch <- prometheus.MustNewConstMetric(
		v.births, prometheus.CounterValue,
		float64(stats.Births),
	)
	ch <- prometheus.MustNewConstMetric(
		v.deaths, prometheus.CounterValue,
		float64(stats.Deaths),
	)
	ch <- prometheus.MustNewConstMetric(
		v.immigrants, prometheus.CounterValue,
		float64(stats.Immigrants),
	)
}
