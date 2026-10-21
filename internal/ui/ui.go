package ui

import (
	"fmt"
	"lifeSimulator/internal/Colonist"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/resources"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func Run(c *colony.Colony, jobs chan<- colony.Job) error {
	snapshot := c.Snapshot()
	app := tview.NewApplication()
	quit := make(chan struct{})
	var quitOnce sync.Once
	requestQuit := func() {
		quitOnce.Do(func() {
			close(quit)
		})
	}
	defer requestQuit()
	list := tview.NewList().ShowSecondaryText(false)
	list.SetBorder(true).SetTitle(" Жители ")
	list.SetBorderColor(tcell.ColorGreen)
	list.SetSelectedBackgroundColor(tcell.ColorDarkSlateGray)

	details := tview.NewTextView().SetDynamicColors(true)
	details.SetBorder(true).SetTitle(" Карточка Жителя ")
	details.SetBorderColor(tcell.ColorGreen)

	bar := func(value int) string {
		n := min(10, max(0, value)/10)
		return strings.Repeat("█", n) + strings.Repeat("░", 10-n)
	}

	events := tview.NewTextView()
	events.SetBorder(true).SetTitle(" События ")
	events.SetBorderColor(tcell.ColorGreen)

	village := tview.NewTextView().SetDynamicColors(true)
	village.SetBorder(true).SetTitle(" Деревня ")
	village.SetBorderColor(tcell.ColorGreen)

	constructionButton := tview.NewButton("Приостановить стройку")
	show := func(index int) {
		snapshot = c.Snapshot()
		if snapshot.ConstructionAllowed {
			constructionButton.SetLabel("Приостановить стройку")
		} else {
			constructionButton.SetLabel("Разрешить стройку")
		}
		for i, col := range snapshot.Colonists {
			label := colonistLabel(col)

			if i >= list.GetItemCount() {
				list.AddItem(label, "", 0, nil)
			} else {
				list.SetItemText(i, label, "")
			}
		}
		total := len(snapshot.Colonists)
		busy := 0
		for _, col := range snapshot.Colonists {
			if col.Busy {
				busy++
			}
		}

		reserveText := "—"
		if total > 0 {
			reserveText = fmt.Sprintf(
				"примерно %.1f мин",
				snapshot.FoodReserveMinutes,
			)
		}

		produced := snapshot.FoodProducedLastMinute
		consumed := snapshot.FoodConsumedLastMinute

		village.SetText(fmt.Sprintf(
			"Жители: [green]%d[-]   Заняты: %d   Свободны: %d\n"+
				"Еда: [yellow]%d[-]   Дерево: [yellow]%d[-]   Дома: %d   Мест: %d   Погода: [cyan]%s[-]\n"+
				"Еда за 60 с: добыто [green]%d[-]   съедено [yellow]%d[-]   баланс: %+d\n"+
				"Запас без новой добычи: %s",
			total, busy, total-busy,
			snapshot.Food, snapshot.Wood,
			snapshot.Houses, snapshot.Capacity, snapshot.Weather,
			produced, consumed, produced-consumed,
			reserveText,
		))
		events.SetText(strings.Join(snapshot.Events, "\n")).ScrollToEnd()
		if index < 0 || index >= len(snapshot.Colonists) {
			details.SetText("Жителей нет")
			return
		}

		col := snapshot.Colonists[index]
		status := col.Action

		progressText := "—"
		if col.ActionDuration > 0 && !col.ActionStartedAt.IsZero() {
			elapsed := time.Since(col.ActionStartedAt)
			progress := int(elapsed * 100 / col.ActionDuration)
			progress = min(100, max(0, progress))

			progressText = fmt.Sprintf(
				"[green]%s[-] %d%%",
				bar(progress),
				progress,
			)
		}

		details.SetText(fmt.Sprintf(
			"%s\nВозраст: %d\n\n"+
				"Действие: %s\n"+
				"Прогресс:  %s\n\n"+
				"Голод:     [yellow]%s[-] %d/100\n"+
				"Усталость: [cyan]%s[-] %d/100\n"+
				"Любимое занятие: %s\n\n"+
				"Рубка дерева: %s\n"+
				"Сбор еды: %s\n\n"+
				"Всего добыто:\n"+
				"Еда: [green]%d[-]\n"+
				"Дерево: [yellow]%d[-]",
			colonistLabel(col), col.Age, status,
			progressText,
			bar(col.Hunger), col.Hunger,
			bar(col.Fatigue), col.Fatigue,
			col.PreferredWork,
			skillText(col.Woodcutting),
			skillText(col.Gathering),
			col.FoodGathered,
			col.WoodGathered,
		))
	}
	for _, col := range snapshot.Colonists {
		list.AddItem(colonistLabel(col), "", 0, nil)
	}

	onSelect := func(
		index int, mainText, secondaryText string, shortcut rune,
	) {
		show(index)
	}

	list.SetChangedFunc(onSelect)
	list.SetSelectedFunc(onSelect)

	notice := tview.NewTextView().SetText("Поручения выполняет любой житель деревни.")

	sendJob := func(resource resources.Resource) {
		select {
		case jobs <- colony.Job{Resource: resource}:
			notice.SetText("Поручение добавлено в очередь")
		default:
			notice.SetText("Очередь заполнена. Попробуйте позже.")
		}
	}

	foodButton := tview.NewButton("Добыть еду").
		SetSelectedFunc(func() {
			sendJob(resources.Food)
		})

	woodButton := tview.NewButton("Добыть дерево").
		SetSelectedFunc(func() {
			sendJob(resources.Wood)
		})

	constructionButton.SetSelectedFunc(func() {
		c.ToggleConstruction()
		show(list.GetCurrentItem())
	})
	foodButton.SetDisabled(jobs == nil)
	woodButton.SetDisabled(jobs == nil)

	buttons := tview.NewGrid().
		SetColumns(0, 24, 0).
		SetRows(0, 1, 1, 1, 1, 1, 0).
		AddItem(woodButton, 1, 1, 1, 1, 0, 0, false).
		AddItem(foodButton, 3, 1, 1, 1, 0, 0, false).
		AddItem(constructionButton, 5, 1, 1, 1, 0, 0, false)

	buttons.SetBorder(true).SetTitle(" Поручения деревне ")
	buttons.SetBorderColor(tcell.ColorGreen)

	content := tview.NewFlex().
		AddItem(list, 22, 0, true).
		AddItem(details, 0, 1, false).
		AddItem(buttons, 0, 1, false)

	layout := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(village, 6, 0, false).
		AddItem(content, 0, 3, true).
		//AddItem(buttons, 0, 1, false).
		AddItem(notice, 3, 0, false).
		AddItem(events, 0, 1, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape ||
			event.Key() == tcell.KeyCtrlC {
			requestQuit()
			return nil
		}

		if event.Rune() == 'r' {
			show(list.GetCurrentItem())
			return nil
		}

		return event
	})

	var startOnce sync.Once

	app.SetAfterDrawFunc(func(_ tcell.Screen) {
		startOnce.Do(func() {
			go func() {
				ticker := time.NewTicker(500 * time.Millisecond)
				defer ticker.Stop()

				for {
					select {
					case <-quit:
						app.Stop()
						return

					case <-ticker.C:
						app.QueueUpdateDraw(func() {
							show(list.GetCurrentItem())
						})
					}
				}
			}()
		})
	})

	return app.SetRoot(layout, true).
		EnableMouse(true).
		SetFocus(list).
		Run()
}

func skillText(skill Colonist.Skill) string {
	level := skill.Level()

	if level == Colonist.MaxSkillLevel {
		return fmt.Sprintf("уровень %d (максимум)", level)
	}

	return fmt.Sprintf(
		"уровень %d, опыт %d/%d",
		level,
		skill.XP%Colonist.XPerLevel,
		Colonist.XPerLevel)
}

func levelColor(level int) string {
	switch level {
	case 2:
		return "lime"
	case 3:
		return "cyan"
	case 4:
		return "yellow"
	case 5:
		return "fuchsia"
	default:
		return "white"
	}
}

func colonistLabel(col Colonist.Colonist) string {
	level := col.Level()

	return fmt.Sprintf(
		"[%s]%s (ур. %d)[-]",
		levelColor(level),
		tview.Escape(col.Name),
		level,
	)
}
