package utilities

import "encoding/json"

type Serializable interface {
	Serialize() ([]byte, error)
}

func Serialize[T any](content any) ([]byte, error) {
	json, err := json.Marshal(content)
	if err != nil {
		return nil, err
	}

	return json, nil
}
