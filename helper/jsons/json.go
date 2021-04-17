package jsons

import "encoding/json"

func ToStringJson(data interface{}) (out string, err error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}
	out = string(jsonData)
	return
}

func ToStringJsonNoError(data interface{}) (out string) {
	out, _ = ToStringJson(data)
	return
}
