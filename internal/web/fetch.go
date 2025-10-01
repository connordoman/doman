package web

import (
	"io"
	"net/http"
)

type RequestInit struct {
	Body string `json:"body"`
}

func Fetch(url string) (string, error) {
	response, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	return string(body), nil
}
