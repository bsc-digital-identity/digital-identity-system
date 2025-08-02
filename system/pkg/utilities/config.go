package utilities

import (
	"encoding/json"
	"os"
)

func ReadConfig[T any](file string) (T, error) {
	var empty T

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return empty, err
	}

	var config T
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return empty, err
	}

	return config, nil
}
