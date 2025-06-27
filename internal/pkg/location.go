package pkg

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const ipInfoUrl = "https://ipinfo.io/json"

// IPInfo.io response
type IPInfoResponse struct {
	City     string `json:"city"`
	Country  string `json:"country"`
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
	Loc      string `json:"loc"` // Location in "lat,lon" format
	Org      string `json:"org"`
	Postal   string `json:"postal"`
	Readme   string `json:"readme"` // API documentation link
	Region   string `json:"region"`
	Timezone string `json:"timezone"`
}

func GetIPInfo() (*IPInfoResponse, error) {
	resp, err := http.Get(ipInfoUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to get current location: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var ipInfo IPInfoResponse
	if err := json.Unmarshal(body, &ipInfo); err != nil {
		return nil, fmt.Errorf("failed to unmarshal IP info: %w", err)
	}

	return &ipInfo, nil
}
