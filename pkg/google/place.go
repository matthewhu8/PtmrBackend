package google

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

type PlaceDetailsResponse struct {
	Name             string   `json:"name"`
	ID               string   `json:"id"`
	Types            []string `json:"types"`
	FormattedAddress string   `json:"formattedAddress"`
	Location         Location `json:"location"`
	Photos           []Photo  `json:"photos"`
	GoogleMapsURI    string   `json:"googleMapsUri"`
	DisplayName      struct {
		Text string `json:"text"`
	} `json:"displayName"`
	NationalPhoneNumber string       `json:"nationalPhoneNumber"`
	PriceLevel          string       `json:"priceLevel,omitempty"`
	Rating              float32      `json:"rating"`
	RegularOpeningHours OpeningHours `json:"regularOpeningHours"`
	WebsiteURI          string       `json:"websiteUri"`
}

type PlaceSearchResponse struct {
	Results []struct {
		PlaceID string `json:"place_id"`
	} `json:"results"`
}

type OpeningHours struct {
	Periods []struct {
		Open  Point `json:"open"`
		Close Point `json:"close"`
	} `json:"periods"`
	WeekdayDescriptions []string `json:"weekdayDescriptions"`
}

type Point struct {
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Photo struct {
	Name     string `json:"name"`
	WidthPx  int    `json:"widthPx"`
	HeightPx int    `json:"heightPx"`
}

func (g *Service) GetPlaceID(query string) (string, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	baseURL := "https://maps.googleapis.com/maps/api/place/textsearch/json"
	params := url.Values{}
	params.Add("query", query)
	params.Add("key", apiKey)
	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var searchResponse PlaceSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return "", err
	}
	fmt.Println(searchResponse.Results)
	if len(searchResponse.Results) > 0 {
		return searchResponse.Results[0].PlaceID, nil
	}
	return "", fmt.Errorf("no results found")
}

func (g *Service) GetPlaceDetails(placeID string) (*PlaceDetailsResponse, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	baseURL := fmt.Sprintf("https://places.googleapis.com/v1/places/%s", url.QueryEscape(placeID))
	fields := "name,id,photos,formattedAddress,location,types,displayName,googleMapsUri,nationalPhoneNumber,priceLevel,rating,regularOpeningHours,websiteUri"
	requestURL := fmt.Sprintf("%s?fields=%s&key=%s", baseURL, fields, apiKey)

	resp, err := http.Get(requestURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var placeDetails PlaceDetailsResponse
	if err := json.NewDecoder(resp.Body).Decode(&placeDetails); err != nil {
		return nil, err
	}

	if placeDetails.Name == "" {
		return nil, fmt.Errorf("failed to get place details")
	}

	return &placeDetails, nil
}
