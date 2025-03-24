package google

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetPlaceID(t *testing.T) {
	if skipAPICalls() {
		t.Skip("Skipping TestGetPlaceID to reduce API calls")
	}
	query := "Chili NYC"
	expectedPlaceID := "ChIJJS3mqONZwokR9KlP3H_7MNg"
	service := NewGoogleService()
	placeID, err := service.GetPlaceID(query)
	require.NoError(t, err)
	require.Equal(t, expectedPlaceID, placeID)
}

func TestGetPlaceDetails(t *testing.T) {
	if skipAPICalls() {
		t.Skip("Skipping TestGetPlaceDetails to reduce API calls")
	}
	placeID := "ChIJJS3mqONZwokR9KlP3H_7MNg"
	service := NewGoogleService()
	details, err := service.GetPlaceDetails(placeID)
	require.NoError(t, err)
	fmt.Println(details.Name)
	fmt.Println(details.ID)
	fmt.Println(details.Location.Latitude)
	fmt.Println(details.Location.Longitude)
	fmt.Println(details.Types)
	for _, photo := range details.Photos {
		fmt.Println(photo.Name)
	}
	fmt.Println(details.GoogleMapsURI)
	fmt.Println(details.DisplayName.Text)
	fmt.Println(details.NationalPhoneNumber)
	fmt.Println(details.PriceLevel)
	fmt.Println(details.Rating)
	for _, day := range details.RegularOpeningHours.Periods {
		fmt.Println(day.Open)
	}
	fmt.Println(details.WebsiteURI)
	fmt.Println(details.FormattedAddress)
}
