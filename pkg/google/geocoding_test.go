package google

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetLatLon(t *testing.T) {
	if skipAPICalls() {
		t.Skip("Skipping TestGetLatLon to reduce API calls")
	}
	service := NewGoogleService()
	lat, lon, err := service.GetLatLon("1600 Amphitheatre Parkway, Mountain View, CA")
	require.NoError(t, err)
	require.NotZero(t, lat)
	require.NotZero(t, lon)
	require.Equal(t, 37.422535, lat)
	require.Equal(t, -122.0847281, lon)
}

func TestGetPlaceIDOfAddress(t *testing.T) {
	if skipAPICalls() {
		t.Skip("Skipping TestGetPlaceID to reduce API calls")
	}
	service := NewGoogleService()
	placeID, err := service.GetPlaceIDOfAddress("1600 Amphitheatre Parkway, Mountain View, CA")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	require.NoError(t, err)
	require.Equal(t, "ChIJF4Yf2Ry7j4AR__1AkytDyAE", placeID)
}
