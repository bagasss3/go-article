package helper

import (
	"encoding/json"
	"reflect"
	"strings"
	"time"
)

func ParseTimeDuration(t string, defaultt time.Duration) time.Duration {
	timeDurr, err := time.ParseDuration(t)
	if err != nil {
		return defaultt
	}
	return timeDurr
}

func ParseTime(t string) (time.Time, error) {
	parsedTime, err := time.Parse("2006-01-02T15:04", t)
	if err != nil {
		return time.Time{}, err
	}
	return parsedTime, nil
}

func GetValueBetween(value string, a string, b string) string {
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func ToJSON(v any) string {
	if v == nil {
		return "{}"
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Interface:
		if rv.IsNil() {
			return "{}"
		}
	}

	b, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(b)
}
