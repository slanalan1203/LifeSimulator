package api

import (
	"encoding/json"
	"io"
	"lifeSimulator/internal/colony"
	"lifeSimulator/internal/resources"
	"net/http"
	"strconv"
)

type villageResponse struct {
	Population             int            `json:"population"`
	Food                   int            `json:"food"`
	Wood                   int            `json:"wood"`
	Houses                 int            `json:"houses"`
	Capacity               int            `json:"capacity"`
	Weather                colony.Weather `json:"weather"`
	ConstructionAllowed    bool           `json:"construction_allowed"`
	FoodProducedLastMinute int            `json:"food_produced_last_minute"`
	FoodConsumedLastMinute int            `json:"food_consumed_last_minute"`
	FoodReserveMinutes     float64        `json:"food_reserve_minutes"`
}

type colonistResponse struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Busy             bool   `json:"busy"`
	Hunger           int    `json:"hunger"`
	Fatigue          int    `json:"fatigue"`
	Action           string `json:"action"`
	WoodcuttingLevel int    `json:"woodcutting"`
	GatheringLevel   int    `json:"gathering"`
}
type colonistsResponse struct {
	Colonists []colonistResponse `json:"colonists"`
}

type jobRequest struct {
	Resource resources.Resource `json:"resource"`
}

func NewHandler(c *colony.Colony, jobs chan<- colony.Job) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/village",
		func(w http.ResponseWriter, r *http.Request) {
			snapshot := c.Snapshot()

			response := villageResponse{
				Population: len(snapshot.Colonists),
				Food:       snapshot.Food,
				Wood:       snapshot.Wood,
				Houses:     snapshot.Houses,
				Capacity:   snapshot.Capacity,
				Weather:    snapshot.Weather,

				ConstructionAllowed:    snapshot.ConstructionAllowed,
				FoodProducedLastMinute: snapshot.FoodProducedLastMinute,
				FoodConsumedLastMinute: snapshot.FoodConsumedLastMinute,
				FoodReserveMinutes:     snapshot.FoodReserveMinutes,
			}

			body, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Не удалось сформировать ответ", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		})

	mux.HandleFunc("GET /api/v1/colonists",
		func(w http.ResponseWriter, r *http.Request) {
			snapshot := c.Snapshot()

			response := colonistsResponse{
				Colonists: make([]colonistResponse, 0, len(snapshot.Colonists)),
			}
			for _, col := range snapshot.Colonists {
				response.Colonists = append(response.Colonists, colonistResponse{
					ID:               col.ID,
					Name:             col.Name,
					Busy:             col.Busy,
					Hunger:           col.Hunger,
					Fatigue:          col.Fatigue,
					Action:           col.Action,
					WoodcuttingLevel: col.Woodcutting.Level(),
					GatheringLevel:   col.Gathering.Level(),
				})
			}

			body, err := json.Marshal(response)
			if err != nil {
				http.Error(w, "Не удалось сформировать ответ", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(body)
		})

	mux.HandleFunc("GET /api/v1/colonists/{id}",
		func(w http.ResponseWriter, r *http.Request) {
			id, err := strconv.Atoi(r.PathValue("id"))
			if err != nil || id < 0 {
				http.Error(w, "некорректный id жителя", http.StatusBadRequest)
				return
			}

			snapshot := c.Snapshot()
			for _, col := range snapshot.Colonists {
				if col.ID != id {
					continue
				}

				response := colonistResponse{
					ID:               col.ID,
					Name:             col.Name,
					Busy:             col.Busy,
					Hunger:           col.Hunger,
					Fatigue:          col.Fatigue,
					Action:           col.Action,
					WoodcuttingLevel: col.Woodcutting.Level(),
					GatheringLevel:   col.Gathering.Level(),
				}

				body, err := json.Marshal(response)
				if err != nil {
					http.Error(w, "не удалось сформировать ответ", http.StatusInternalServerError)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(body)
				return
			}
			http.Error(w, "Житель не найден", http.StatusNotFound)
		})

	mux.HandleFunc("POST /api/v1/jobs", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 1024)
		defer r.Body.Close()

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		var job jobRequest
		if err := decoder.Decode(&job); err != nil {
			http.Error(w, "Неккоректный JSON", http.StatusBadRequest)
			return
		}

		var extra any
		if err := decoder.Decode(&extra); err != io.EOF {
			http.Error(w, "Ожидается один JSON", http.StatusBadRequest)
			return
		}

		if job.Resource != resources.Wood && job.Resource != resources.Food {
			http.Error(w, "Ресурс должен быть Wood или Food", http.StatusBadRequest)
			return
		}
		select {
		case jobs <- colony.Job{Resource: job.Resource}:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			_, _ = w.Write([]byte(`{"status":"queued"}`))
		default:
			http.Error(w, "Очередь поручений заполнена", http.StatusServiceUnavailable)
		}
	})
	return mux
}
