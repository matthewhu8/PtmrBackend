package google

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type GeocodingResponse struct {
	Results []struct {
		Geometry struct {
			Location struct {
				Lat float64 `json:"lat"`
				Lng float64 `json:"lng"`
			} `json:"location"`
		} `json:"geometry"`
		PlaceID string `json:"place_id"`
	} `json:"results"`
	Status string `json:"status"`
}

func (g *Service) GetLatLon(address string) (float64, float64, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"
	encodedAddress := url.QueryEscape(address)
	requestURL := fmt.Sprintf("%s?address=%s&key=%s", baseURL, encodedAddress, apiKey)

	resp, err := http.Get(requestURL)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var geoResp GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return 0, 0, err
	}

	if geoResp.Status != "OK" || len(geoResp.Results) == 0 {
		return 0, 0, fmt.Errorf("failed to get coordinates for address: %s", address)
	}

	lat := geoResp.Results[0].Geometry.Location.Lat
	lon := geoResp.Results[0].Geometry.Location.Lng

	return lat, lon, nil
}

func (g *Service) GetPlaceIDOfAddress(address string) (string, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	baseURL := "https://maps.googleapis.com/maps/api/geocode/json"

	encodedAddress := url.QueryEscape(address)
	requestURL := fmt.Sprintf("%s?address=%s&key=%s", baseURL, encodedAddress, apiKey)

	resp, err := http.Get(requestURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var geoResp GeocodingResponse
	if err := json.NewDecoder(resp.Body).Decode(&geoResp); err != nil {
		return "", err
	}

	if geoResp.Status != "OK" || len(geoResp.Results) == 0 {
		return "", fmt.Errorf("failed to get place_id for address: %s", address)
	}

	placeID := geoResp.Results[0].PlaceID

	return placeID, nil
}
