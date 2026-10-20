package namegen

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func Random() (string, error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(
		"https://randomuser.me/api/1.4/?inc=name&nat=us,gb&noinfo",
	)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Сервис имен: %s", resp.Status)
	}

	var data struct {
		Results []struct {
			Name struct {
				First string `json:"first"`
			} `json:"name"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "", err
	}

	if len(data.Results) == 0 {
		return "", fmt.Errorf("Сервис не вернул имя")
	}

	name := strings.TrimSpace(data.Results[0].Name.First)
	if name == "" {
		return "", fmt.Errorf("Сервис вернкл пустое имя")
	}
	return name, nil
}
