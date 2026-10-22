package namegen

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const batchSize = 100

type Pool struct {
	names  chan string
	client *http.Client
}

func NewPool() *Pool {
	return &Pool{
		names: make(chan string, 2*batchSize),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (p *Pool) Next() string {
	select {
	case name := <-p.names:
		return name
	default:
		return ""
	}
}

func (p *Pool) Run(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		if ctx.Err() != nil {
			return
		}
		if len(p.names) < batchSize {
			names, err := p.fetch(ctx)
			if err == nil {
				for _, name := range names {
					select {
					case <-ctx.Done():
						return
					case p.names <- name:
					}
				}
			}
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (p *Pool) fetch(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("https://randomuser.me/api/1.4/?results=%d&inc=name&nat=us,gb&noinfo", batchSize)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("сервис имён: %s", resp.Status)
	}

	var data struct {
		Results []struct {
			Name struct {
				First string `json:"first"`
			} `json:"name"`
		} `json:"results"`
	}

	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&data); err != nil {
		return nil, err
	}

	names := make([]string, 0, batchSize)
	for _, result := range data.Results {
		if name := strings.TrimSpace(result.Name.First); name != "" {
			names = append(names, name)
			if len(names) == batchSize {
				break
			}
		}
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("сервис не вернул имена")
	}
	return names, nil
}
