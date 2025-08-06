package utilities

import (
	"encoding/json"
	"os"
)

type JsonConfigObj[T any] interface {
	ConvertToDomain() T
}

func ReadConfig[T JsonConfigObj[U], U any](file string) (U, error) {
	var empty U

	fileContent, err := os.ReadFile(file)
	if err != nil {
		return empty, err
	}

	var config T
	err = json.Unmarshal(fileContent, &config)
	if err != nil {
		return empty, err
	}

	return config.ConvertToDomain(), nil
}

func ConvertJsonArrayToDomain[T JsonConfigObj[U], U any](jsonArray []T) []U {
	var domainArray []U
	for _, item := range jsonArray {
		domainArray = append(domainArray, item.ConvertToDomain())
	}
	return domainArray
}
