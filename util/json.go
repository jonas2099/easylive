package util

import "encoding/json"

func JSON(i interface{}) string {
	data, _ := json.Marshal(i)
	return string(data)
}
