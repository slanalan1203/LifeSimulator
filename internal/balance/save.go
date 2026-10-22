package balance

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("неверные настройки: %w", err)
	}

	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("подготовить настройки: %w", err)
	}
	content = append(content, '\n')

	temporary, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("создать временный файл настроек: %w", err)
	}
	defer func() {
		temporary.Close()
		os.Remove(temporary.Name())
	}()

	if _, err := temporary.Write(content); err != nil {
		return fmt.Errorf("записать настройки: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("закрыть файл настроек: %w", err)
	}
	if err := os.Rename(temporary.Name(), path); err != nil {
		return fmt.Errorf("сохранить настройки: %w", err)
	}

	return nil
}
