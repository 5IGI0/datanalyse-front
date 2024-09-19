package main

import (
	"encoding/json"
	"net/http"
)

func ApiEndpoint(handler func(*http.Request) (any, error)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		response, err := handler(r)

		var error_string *string

		if err != nil {
			a := err.Error()
			error_string = &a
		}

		data, err := json.Marshal(map[string]any{
			"success": err == nil,
			"error":   error_string,
			"data":    response})
		AssertError(err)
		w.Header().Add("Content-Type", "application/json")
		w.Write(data)
	}
}
