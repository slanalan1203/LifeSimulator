package colony

import (
	"context"
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/balance"
	"lifeSimulator/internal/resources"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

const (
	residentPerHouse = 5
	houseWoodCost    = 10
)

type Colony struct {
	mu        sync.Mutex
	colonists []Colonist.Colonist
	storage   map[resources.Resource]int

	houses int
	//hasHouse bool

	eventsMu sync.Mutex
	events   []string

	weather Weather

	restockFood bool
	restockWood bool

	buildingHouse bool

	foodProduced int
	foodConsumed int
	births       uint64
	deaths       uint64
	immigrants   uint64

	foodHistory []foodChange

	allowConstruction bool

	cfg balance.Config

	needsChanged chan struct{}
	yearChanged  chan struct{}
}

type Snapshot struct {
	Colonists []Colonist.Colonist
	Food      int
	Wood      int
	Houses    int
	Capacity  int

	Events []string

	Weather Weather

	FoodProduced int
	FoodConsumed int

	FoodProducedLastMinute int
	FoodConsumedLastMinute int
	FoodReserveMinutes     float64

	ConstructionAllowed bool

	Population int
	AdultAge   int
}

type Job struct {
	Resource resources.Resource
}

func NewColony(cols []Colonist.Colonist, cfg balance.Config) *Colony {
	colonists := append([]Colonist.Colonist{}, cols...)
	return &Colony{
		storage:   make(map[resources.Resource]int),
		colonists: colonists,

		weather: WeatherClear,

		allowConstruction: true,
		cfg:               cfg,
		needsChanged:      make(chan struct{}, 1),
		yearChanged:       make(chan struct{}, 1),
	}
}

func (c *Colony) DoJob(ctx context.Context, workerIndex int, resource resources.Resource) {
	if err := ctx.Err(); err != nil {
		return
	}
	c.mu.Lock()

	col := &c.colonists[workerIndex]
	if col.Dead || col.Age < c.cfg.Population.AdultAge {
		c.mu.Unlock()
		return
	}
	col.Busy = true
	col.EmployType = resource
	name := col.Name
	needsRest := col.Fatigue >= 75
	restDuration := 3 * time.Second
	if c.houses > 0 {
		restDuration = 2 * time.Second
	}
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		defer c.mu.Unlock()

		col := &c.colonists[workerIndex]
		if col.Dead {
			col.Action = "умер"
		} else {
			col.Action = "свободен"
		}
		col.Busy = false
		col.EmployType = resources.NotSet

		col.ActionStartedAt = time.Time{}
		col.ActionDuration = 0
	}()

	if needsRest {
		c.logf("%s устал и пошел отдыхать\n", name)
		c.startAction(workerIndex, "отдыхает", restDuration)
		if err := waitForAction(ctx, restDuration); err != nil {
			c.logf("%s прервал работу", name)
			return
		}
		c.mu.Lock()
		col = &c.colonists[workerIndex]
		if col.Dead {
			c.mu.Unlock()
			return
		}
		col.Fatigue = 0
		c.mu.Unlock()
		c.logf("%s отдохнул \n", name)
	}
	c.Eat(workerIndex)

	c.mu.Lock()
	col = &c.colonists[workerIndex]
	if col.Dead {
		c.mu.Unlock()
		return
	}
	isHungry := col.Hunger >= c.cfg.Needs.HungerPerMeal
	isRaining := c.weather == WeatherRain
	work := c.cfg.Work
	workLevel := col.Gathering.Level()
	if resource == resources.Wood {
		workLevel = col.Woodcutting.Level()
	}
	c.mu.Unlock()

	workDuration := 2 * time.Second

	switch resource {
	case resources.Wood:
		workDuration = time.Duration(work.WoodSeconds) * time.Second
	case resources.Food:
		workDuration = time.Duration(work.FoodSeconds) * time.Second
	}

	if isHungry {
		workDuration *= 2
		c.logf("Очень голоден, но все равно работаю %s\n", name)
	}

	if isRaining && resource == resources.Wood {
		workDuration *= 2
		c.logf("%s: дождь замедляет рубку дерева", name)
	}

	durationPercent := 100 - (workLevel-1)*10
	workDuration = workDuration * time.Duration(durationPercent) / 100

	action := "собирает еду"

	if resource == resources.Wood {
		action = "рубит дерево"
	}
	c.startAction(workerIndex, action, workDuration)
	c.logf("Начал работу: %s\n", name)

	err := waitForAction(ctx, workDuration)
	if err != nil {
		c.logf("%s прервал работу", name)
		return
	}

	amount := 1
	if rand.IntN(100) < 20 {
		amount = 2
	}
	if isRaining && resource == resources.Food {
		amount++
	}
	c.mu.Lock()
	if ctx.Err() != nil {
		c.mu.Unlock()
		return
	}

	col = &c.colonists[workerIndex]
	if col.Dead {
		c.mu.Unlock()
		return
	}
	fatigueGain := 25
	if resource == col.PreferredWork {
		fatigueGain = 15
	}
	col.Fatigue = min(100, col.Fatigue+fatigueGain)
	c.storage[resource] += amount
	switch resource {
	case resources.Food:
		col.Gathering.AddXp()
		col.FoodGathered += amount
		c.recordFoodChangedLocked(amount, 0)

	case resources.Wood:
		col.Woodcutting.AddXp()
		col.WoodGathered += amount
	}
	c.mu.Unlock()
	c.logf("%s закончил работу: %s +%d\n", name, resource, amount)
}

func (c *Colony) Work(ctx context.Context, workerIndex int, jobs <-chan Job) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			c.DoJob(ctx, workerIndex, job.Resource)
		}
	}
}

func (c *Colony) Live(ctx context.Context, workerIndex int, jobs <-chan Job) {
	stop := ctx.Done()
	for {
		select {
		case <-stop:
			return
		default:
		}
		c.mu.Lock()
		dead := c.colonists[workerIndex].Dead
		isChild := c.colonists[workerIndex].Age < c.cfg.Population.AdultAge
		c.mu.Unlock()
		if dead {
			return
		}
		if isChild {
			c.startAction(workerIndex, "играет", 2*time.Second)

			if err := waitForAction(ctx, 2*time.Second); err != nil {
				return
			}
			continue
		}
		select {
		case job, ok := <-jobs:
			if ok {
				c.DoJob(ctx, workerIndex, job.Resource)
				continue
			}
			jobs = nil
		default:
		}
		if c.TryBuildHouse(workerIndex, stop) {
			continue
		}
		resource := c.chooseWork(workerIndex)

		if resource == resources.NotSet {
			c.startAction(workerIndex, "гуляет", 2*time.Second)

			select {
			case <-stop:
				return
			case job, ok := <-jobs:
				if ok {
					c.DoJob(ctx, workerIndex, job.Resource)
				} else {
					jobs = nil
				}
			case <-time.After(2 * time.Second):
				c.mu.Lock()
				col := &c.colonists[workerIndex]
				if !col.Dead {
					col.Fatigue = max(0, col.Fatigue-5)
				}
				c.mu.Unlock()
			}
			continue
		}
		c.DoJob(ctx, workerIndex, resource)
	}
}

func (c *Colony) Assign(resource resources.Resource) bool {
	if resource != resources.Wood && resource != resources.Food {
		return false
	}
	for i := range c.colonists {

		if c.colonists[i].Busy == true {
			continue
		}
		c.colonists[i].Busy = true
		c.colonists[i].EmployType = resource
		return true
	}
	return false
}

func (c *Colony) Status() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, colonist := range c.colonists {
		status := "свободен"
		if colonist.Busy {
			status = "занят"
		}

		c.logf("%s — %s, голод: %d, усталость: %d\n",
			colonist.Name,
			status,
			colonist.Hunger,
			colonist.Fatigue,
		)
	}
	for k, v := range c.storage {
		fmt.Println("Ресурс:", k, v)
	}
	fmt.Println("Домов:", c.houses)
}

func (c *Colony) Gather(ctx context.Context, resource resources.Resource) bool {
	if resource != resources.Wood && resource != resources.Food {
		return false
	}
	c.mu.Lock()
	for i := range c.colonists {
		if c.colonists[i].Busy {
			continue
		}
		c.colonists[i].EmployType = resource
		c.colonists[i].Busy = true
		name := c.colonists[i].Name
		c.mu.Unlock()
		c.logf("Начал работу %s\n", name)
		c.startAction(i, "работа", 2*time.Second)
		c.logf("Закончил работу %s\n", name)
		c.mu.Lock()
		c.storage[resource]++
		if resource == resources.Food {
			c.recordFoodChangedLocked(1, 0)
		}
		c.colonists[i].Busy = false
		c.mu.Unlock()
		return true
	}
	c.mu.Unlock()
	return false
}

func (c *Colony) Eat(workerIndex int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	colonist := &c.colonists[workerIndex]

	if colonist.Dead || colonist.Hunger < c.cfg.Needs.HungerPerMeal {
		return
	}

	if c.storage[resources.Food] <= 0 {
		c.logf("%s голоден, но на складе нет еды \n", colonist.Name)
		return
	}
	c.storage[resources.Food]--
	c.recordFoodChangedLocked(0, 1)
	colonist.Hunger -= c.cfg.Needs.HungerPerMeal

	c.logf("%s поел, голод %d\n", colonist.Name, colonist.Hunger)
}

func (c *Colony) TryBuildHouse(workerIndex int, stop <-chan struct{}) bool {
	duration := 30 * time.Second
	c.mu.Lock()
	people := c.aliveCountLocked()
	freePlaces := c.houses*residentPerHouse - people
	col := &c.colonists[workerIndex]

	if col.Dead ||
		col.Age < c.cfg.Population.AdultAge ||
		!c.allowConstruction ||
		c.buildingHouse ||
		freePlaces > 1 ||
		c.storage[resources.Food] < c.foodReserveLocked(2*time.Minute) ||
		c.storage[resources.Wood] < houseWoodCost ||
		col.Busy ||
		col.Fatigue >= 75 ||
		col.Hunger >= c.cfg.Needs.HungerPerMeal {
		c.mu.Unlock()
		return false
	}
	c.buildingHouse = true
	c.storage[resources.Wood] -= houseWoodCost

	col.Busy = true
	col.EmployType = resources.NotSet
	col.Action = "строит дом"
	col.ActionStartedAt = time.Now()
	col.ActionDuration = duration
	name := col.Name
	c.mu.Unlock()

	c.logf("%s начал строить дом", name)

	timer := time.NewTimer(duration)
	defer timer.Stop()

	finished := false

	select {
	case <-stop:
	case <-timer.C:
		finished = true
	}
	c.mu.Lock()

	col = &c.colonists[workerIndex]
	finished = finished && !col.Dead
	if finished {
		c.houses++
		col.Fatigue = min(100, col.Fatigue+25)
	} else {
		c.storage[resources.Wood] += houseWoodCost
	}

	houseNumber := c.houses
	c.buildingHouse = false

	col.Busy = false
	if col.Dead {
		col.Action = "умер"
	} else {
		col.Action = "свободен"
	}
	col.ActionStartedAt = time.Time{}
	col.ActionDuration = 0
	c.mu.Unlock()

	if finished {
		c.logf("%s построил дом №%d", name, houseNumber)
	} else {
		c.logf("%s прервал стройку, дерево возвращено", name)
	}

	return true
}

func (c *Colony) ToggleConstruction() {
	c.mu.Lock()

	c.allowConstruction = !c.allowConstruction
	allowed := c.allowConstruction

	c.mu.Unlock()

	if allowed {
		c.logf("Новое строительство разрешено")
	} else {
		c.logf("Новое строительство приостановлено")
	}
}

func (c *Colony) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.snapshotLocked()
}

func (c *Colony) snapshotLocked() Snapshot {
	producedLastMinute := 0
	consumedLastMinute := 0
	cutoff := time.Now().Add(-time.Minute)

	for _, change := range c.foodHistory {
		if change.at.After(cutoff) {
			producedLastMinute += change.produced
			consumedLastMinute += change.consumed
		}
	}

	population := c.aliveCountLocked()
	foodPerMinute := c.foodConsumptionPerMinuteLocked()
	reserveMinutes := 0.0

	if foodPerMinute > 0 {
		reserveMinutes = float64(c.storage[resources.Food]) / foodPerMinute
	}

	c.eventsMu.Lock()
	defer c.eventsMu.Unlock()

	return Snapshot{
		Population: population,
		AdultAge:   c.cfg.Population.AdultAge,
		Colonists:  append([]Colonist.Colonist{}, c.colonists...),
		Food:       c.storage[resources.Food],
		Wood:       c.storage[resources.Wood],
		Houses:     c.houses,
		Capacity:   c.houses * residentPerHouse,

		Events: append([]string{}, c.events...),

		Weather: c.weather,

		FoodConsumed: c.foodConsumed,
		FoodProduced: c.foodProduced,

		FoodConsumedLastMinute: consumedLastMinute,
		FoodProducedLastMinute: producedLastMinute,
		FoodReserveMinutes:     reserveMinutes,

		ConstructionAllowed: c.allowConstruction,
	}
}

func (c *Colony) AddEvent(message string) {
	c.logf("%s", message)
}

func (c *Colony) logf(format string, args ...any) {
	message := strings.TrimSpace(fmt.Sprintf(format, args...))

	c.eventsMu.Lock()
	defer c.eventsMu.Unlock()

	c.events = append(c.events, message)
	if len(c.events) > 100 {
		c.events = c.events[len(c.events)-100:]
	}
}

func (c *Colony) startAction(workerIndex int, action string, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	col := &c.colonists[workerIndex]
	if col.Dead {
		return
	}
	col.Action = action
	col.ActionDuration = duration
	col.ActionStartedAt = time.Now()
}

func (c *Colony) TryAddColonist(name string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	capacity := c.houses * residentPerHouse
	if c.aliveCountLocked() >= capacity || c.storage[resources.Food] < 10 {
		return -1, false
	}

	index := len(c.colonists)
	newColonist := Colonist.NewColonist(index, name, c.cfg.Population.AdultAge)

	if rand.IntN(2) == 0 {
		newColonist.PreferredWork = resources.Food
	}

	c.colonists = append(c.colonists, newColonist)
	c.immigrants++
	c.logf("В деревне новый житель: %s", newColonist.Name)

	return index, true
}
