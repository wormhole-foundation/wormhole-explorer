package commons

import (
	"encoding/json"
	"net/http"
)

func ToJson(headers http.Header) string {
	bytes, _ := json.Marshal(headers)
	return string(bytes)
}

type HttpDo interface {
	Do(req *http.Request) (*http.Response, error)
}
