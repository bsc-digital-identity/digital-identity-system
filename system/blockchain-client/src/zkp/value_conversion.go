package zkp

import (
	"fmt"
	"strconv"
	"time"
)

func convertToVariable(field FieldDefinition, value interface{}) (interface{}, error) {
	if value == nil {
		return nil, fmt.Errorf("value for field '%s' is nil", field.Name)
	}

	switch field.Type {
	case FieldTypeInteger, FieldTypeNumber:
		return convertToInt(value)
	case FieldTypeBoolean:
		return convertToBool(value)
	case FieldTypeString:
		return fmt.Sprint(value), nil
	case FieldTypeDate:
		switch v := value.(type) {
		case time.Time:
			return v.Unix(), nil
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t.Unix(), nil
			}
			if t, err := time.Parse("2006-01-02", v); err == nil {
				return t.Unix(), nil
			}
			return fmt.Sprint(value), nil
		default:
			return fmt.Sprint(value), nil
		}
	default:
		return value, nil
	}
}

func convertToInt(value interface{}) (int64, error) {
	switch v := value.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case string:
		if v == "" {
			return 0, fmt.Errorf("empty string cannot be converted to integer")
		}
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return 0, fmt.Errorf("unable to parse integer: %w", err)
		}
		return int64(parsed), nil
	case fmt.Stringer:
		parsed, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0, fmt.Errorf("unable to parse integer: %w", err)
		}
		return int64(parsed), nil
	default:
		return 0, fmt.Errorf("unsupported numeric type %T", value)
	}
}

func convertToBool(value interface{}) (int64, error) {
	switch v := value.(type) {
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case string:
		parsed, err := strconv.ParseBool(v)
		if err != nil {
			return 0, fmt.Errorf("unable to parse boolean: %w", err)
		}
		if parsed {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("unsupported boolean type %T", value)
	}
}
