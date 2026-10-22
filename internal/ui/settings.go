package ui

import (
	"fmt"
	"lifeSimulator/internal/balance"
	"math"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type settingsMessage struct {
	*tview.TextView
}

func (m *settingsMessage) SetFormAttributes(_ int, labelColor, bgColor, fieldTextColor, fieldBgColor tcell.Color) tview.FormItem {
	m.TextView.SetFormAttributes(0, labelColor, bgColor, fieldTextColor, fieldBgColor)
	return m
}

func newSettingsForm(cfg balance.Config, apply func(balance.Config) error, cancel func()) *tview.Form {
	form := tview.NewForm().
		SetItemPadding(0).
		SetLabelColor(tcell.ColorWhite).
		SetFieldBackgroundColor(tcell.ColorDarkSlateGray).
		SetFieldTextColor(tcell.ColorWhite).
		SetButtonsAlign(tview.AlignCenter).
		SetCancelFunc(cancel)
	form.SetBorder(true).SetTitle(" Настройки деревни ")
	form.SetBorderColor(tcell.ColorGreen)

	candidate := cfg
	type integerField struct {
		label  string
		input  *tview.InputField
		target *int
	}
	var integers []integerField
	addInteger := func(label string, target *int) {
		form.AddInputField(label, strconv.Itoa(*target), 12, tview.InputFieldInteger, nil)
		input := form.GetFormItem(form.GetFormItemCount() - 1).(*tview.InputField)
		integers = append(integers, integerField{label: label, input: input, target: target})
	}

	addInteger("Год (реальных секунд)", &candidate.Time.YearSeconds)
	addInteger("Взросление (лет)", &candidate.Population.AdultAge)
	addInteger("Рождение детей до (лет)", &candidate.Population.FertilityMaxAge)
	form.AddInputField(
		"Шанс рождения на пару в год (0..1)",
		strconv.FormatFloat(candidate.Population.BirthChancePerPairPerYear, 'f', -1, 64),
		12, tview.InputFieldFloat, nil,
	)
	chanceInput := form.GetFormItem(form.GetFormItemCount() - 1).(*tview.InputField)
	addInteger("Еды при рождении", &candidate.Population.BirthFoodCost)
	addInteger("Шаг голода (реальных секунд)", &candidate.Needs.TickSeconds)
	addInteger("Голод взрослого за шаг", &candidate.Needs.AdultHungerPerTick)
	addInteger("Голод ребёнка за шаг", &candidate.Needs.ChildHungerPerTick)
	addInteger("Порция снимает голода", &candidate.Needs.HungerPerMeal)
	addInteger("Добыча дерева (реальных секунд)", &candidate.Work.WoodSeconds)
	addInteger("Добыча еды (реальных секунд)", &candidate.Work.FoodSeconds)
	status := &settingsMessage{tview.NewTextView().
		SetSize(3, 0).
		SetDynamicColors(true).
		SetScrollable(false).
		SetText("Tab — следующее поле. Esc — отмена.\nШанс 0.12 означает 12% на пару за игровой год.")}
	form.AddFormItem(status)

	settingNames := strings.NewReplacer(
		"time.year_seconds", "Длительность года",
		"population.adult_age", "Возраст взросления",
		"population.fertility_max_age", "Возраст рождения детей",
		"fertility_max_age", "Возраст рождения детей",
		"adult_age", "возраст взросления",
		"population.birth_chance_per_pair_per_year", "Шанс рождения",
		"birth_chance_per_pair_per_year", "Шанс рождения",
		"population.birth_food_cost", "Расход еды при рождении",
		"birth_food_cost", "Расход еды при рождении",
		"needs.tick_seconds", "Шаг голода",
		"needs.hunger_per_meal", "Размер порции",
		"needs.adult_hunger_per_tick", "Голод взрослого за шаг",
		"needs.child_hunger_per_tick", "Голод ребёнка за шаг",
		"work.wood_seconds", "Длительность добычи дерева",
		"work.food_seconds", "Длительность добычи еды",
	)
	showError := func(err error) {
		status.SetText("[red]" + tview.Escape(settingNames.Replace(err.Error())) + "[-]")
	}

	form.AddButton("Применить", func() {
		for _, field := range integers {
			value, err := strconv.Atoi(strings.TrimSpace(field.input.GetText()))
			if err != nil {
				showError(fmt.Errorf("%s: введите целое число", field.label))
				return
			}
			*field.target = value
		}
		chance, err := strconv.ParseFloat(strings.TrimSpace(chanceInput.GetText()), 64)
		if err != nil || math.IsNaN(chance) || math.IsInf(chance, 0) {
			showError(fmt.Errorf("Шанс рождения: введите число от 0 до 1, например 0.12"))
			return
		}
		candidate.Population.BirthChancePerPairPerYear = chance
		if err := candidate.Validate(); err != nil {
			showError(err)
			return
		}
		if err := apply(candidate); err != nil {
			showError(err)
		}
	})
	form.AddButton("Отмена", cancel)
	return form
}
