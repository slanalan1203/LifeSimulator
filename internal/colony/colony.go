package colony

import (
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/resources"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

const (
	residentPerHouse = 5
	housWoodCost     = 5
)

type Colony struct {
	mu        sync.Mutex
	colonists []Colonist.Colonist
	storage   map[resources.Resource]int

	houses   int
	hasHouse bool

	eventsMu sync.Mutex
	events   []string

	weather Weather
}

type Snapshot struct {
	Colonists []Colonist.Colonist
	Food      int
	Wood      int
	HasHouse  bool

	Events []string

	Weather Weather
}

type Job struct {
	Resource resources.Resource
}

func NewColony(cols []Colonist.Colonist) *Colony {
	colonists := append([]Colonist.Colonist{}, cols...)
	return &Colony{
		storage:   make(map[resources.Resource]int),
		colonists: colonists,

		weather: WeatherClear,
	}
}

func (c *Colony) DoJob(workerIndex int, resource resources.Resource) {
	c.mu.Lock()

	col := &c.colonists[workerIndex]

	col.Busy = true
	col.EmployType = resource
	name := col.Name
	needsRest := col.Fatigue >= 75
	restDuration := 3 * time.Second
	if c.hasHouse {
		restDuration = 2 * time.Second
	}
	c.mu.Unlock()
	if needsRest {
		c.logf("%s устал и пошел отдыхать\n", name)
		c.startAction(workerIndex, "отдыхает", restDuration)
		time.Sleep(restDuration)
		c.mu.Lock()
		col = &c.colonists[workerIndex]
		col.Fatigue = 0
		c.mu.Unlock()
		c.logf("%s отдохнул \n", name)
	}
	c.Eat(workerIndex)

	c.mu.Lock()
	col = &c.colonists[workerIndex]

	isHungry := col.Hunger >= 60
	isRaining := c.weather == WeatherRain
	c.mu.Unlock()

	workDuration := 2 * time.Second

	if isHungry {
		workDuration = 4 * time.Second
		c.logf("Очень голоден, но все равно работаю %s\n", name)
	}

	if isRaining && resource == resources.Wood {
		workDuration *= 2
		c.logf("%s: дождь замедляет рубку дерева", name)
	}

	action := "собирает еду"

	if resource == resources.Wood {
		action = "рубит дерево"
	}
	c.startAction(workerIndex, action, workDuration)
	c.logf("Начал работу: %s\n", name)

	time.Sleep(workDuration)

	amount := 1
	if rand.IntN(100) < 20 {
		amount = 2
	}
	if isRaining && resource == resources.Food {
		amount++
	}
	c.mu.Lock()
	col = &c.colonists[workerIndex]

	fatigueGain := 25
	if resource == col.PreferredWork {
		fatigueGain = 15
	}
	col.Hunger = min(100, col.Hunger+20)
	col.Fatigue = min(100, col.Fatigue+fatigueGain)
	c.storage[resource] += amount
	col.Busy = false

	col.Action = "свободен"
	col.ActionStartedAt = time.Time{}
	col.ActionDuration = 0

	c.mu.Unlock()

	c.TryBuildHouse()
	c.logf("%s закончил работу: %s +%d\n", name, resource, amount)
}

func (c *Colony) Work(workerIndex int, jobs <-chan Job) {
	for job := range jobs {
		c.DoJob(workerIndex, job.Resource)
	}
}

func (c *Colony) Live(workerIndex int, jobs <-chan Job, stop <-chan struct{}) {
	for {
		select {
		case <-stop:
			return
		default:
		}
		select {
		case job, ok := <-jobs:
			if ok {
				c.DoJob(workerIndex, job.Resource)
				continue
			}
			jobs = nil
		default:
		}
		c.mu.Lock()
		food := c.storage[resources.Food]
		targetFood := len(c.colonists) * 2
		name := c.colonists[workerIndex].Name
		preferredWork := c.colonists[workerIndex].PreferredWork

		c.mu.Unlock()

		resource := preferredWork
		if food < targetFood {
			resource = resources.Food
		}

		c.logf("%s решил добывать %s\n", name, resource)
		c.DoJob(workerIndex, resource)
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
	fmt.Println("Дом построен:", c.hasHouse)
}

func (c *Colony) Gather(resource resources.Resource) bool {
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
		time.Sleep(2 * time.Second)
		c.logf("Закончил работу %s\n", name)
		c.mu.Lock()
		c.storage[resource]++
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

	if colonist.Hunger < 60 {
		return
	}

	if c.storage[resources.Food] <= 0 {
		c.logf("%s голоден, но на складе нет еды \n", colonist.Name)
		return
	}
	c.storage[resources.Food]--
	colonist.Hunger -= 60

	c.logf("%s поел, голод %d\n", colonist.Name, colonist.Hunger)
}

func (c *Colony) TryBuildHouse() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.hasHouse || c.storage[resources.Wood] < 5 {
		return
	}
	c.storage[resources.Wood] -= 5
	c.hasHouse = true
	c.logf("Жители построили общий дом! Потрачено 5 дерева.\n")
}

func (c *Colony) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.eventsMu.Lock()
	defer c.eventsMu.Unlock()

	return Snapshot{
		Colonists: append([]Colonist.Colonist{}, c.colonists...),
		Food:      c.storage[resources.Food],
		Wood:      c.storage[resources.Wood],
		HasHouse:  c.hasHouse,

		Events: append([]string{}, c.events...),

		Weather: c.weather,
	}
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
	col.Action = action
	col.ActionDuration = duration
	col.ActionStartedAt = time.Now()
}

func (c *Colony) TryAddColonist(name string) (int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.hasHouse || len(c.colonists) >= 5 || c.storage[resources.Food] < 10 {
		return -1, false
	}

	index := len(c.colonists)
	newColonist := Colonist.NewColonist(index, name)

	if rand.IntN(2) == 0 {
		newColonist.PreferredWork = resources.Food
	}

	c.colonists = append(c.colonists, newColonist)
	c.logf("В деревне новый житель: %s", name)

	return index, true
}
