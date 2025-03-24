package util

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz"

// Create a new rand.Rand instance with a seed
var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// RandomInt generates a random integer between min and max
func RandomInt(min, max int64) int64 {
	return min + seededRand.Int63n(max-min+1)
}

func RandomRating() int {
	return seededRand.Intn(100)
}

// RandomString generates a random string of length n
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[seededRand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// RandomEmail generates a random email
func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}

// RandomPhoneNumber generates a random phone number
func RandomPhoneNumber() string {
	return fmt.Sprintf("%d-%03d-%04d", RandomInt(100, 999), RandomInt(100, 999), RandomInt(1000, 9999))
}

// RandomBool generates a random boolean value
func RandomBool() bool {
	return seededRand.Intn(2) == 1
}

// Availability represents the availability for a day
type Availability struct {
	Morning   bool `json:"morning"`
	Afternoon bool `json:"afternoon"`
	Evening   bool `json:"evening"`
	Night     bool `json:"night"`
}

// RandomAvailability generates random availability for each day of the week
// and returns it as a JSON-formatted []byte slice containing JSON strings.
func RandomAvailability() []byte {
	var availabilities []string
	for i := 0; i < 7; i++ {
		availability := Availability{
			Morning:   RandomBool(),
			Afternoon: RandomBool(),
			Evening:   RandomBool(),
			Night:     RandomBool(),
		}
		jsonData, _ := json.Marshal(availability)
		availabilities = append(availabilities, string(jsonData))
	}

	result, _ := json.Marshal(availabilities)

	return result
}

var streetNames = []string{
	"Main St", "Broadway", "Elm St", "Maple Ave", "Oak St",
	"Pine St", "Cedar St", "Birch St", "Walnut St", "2nd St",
}

var cities = []string{
	"New York", "Los Angeles", "Chicago", "Houston", "Phoenix",
	"Philadelphia", "San Antonio", "San Diego", "Dallas", "San Jose",
}

var states = []string{
	"AL", "AK", "AZ", "AR", "CA", "CO", "CT", "DE", "FL", "GA",
	"HI", "ID", "IL", "IN", "IA", "KS", "KY", "LA", "ME", "MD",
	"MA", "MI", "MN", "MS", "MO", "MT", "NE", "NV", "NH", "NJ",
	"NM", "NY", "NC", "ND", "OH", "OK", "OR", "PA", "RI", "SC",
	"SD", "TN", "TX", "UT", "VT", "VA", "WA", "WV", "WI", "WY",
}

func randomZipCode() string {
	return fmt.Sprintf("%05d", rand.Intn(100000))
}

func randomStreetNumber() int {
	return rand.Intn(9999) + 1
}

func RandomUSAddress() string {
	streetNumber := randomStreetNumber()
	streetName := streetNames[rand.Intn(len(streetNames))]
	city := cities[rand.Intn(len(cities))]
	state := states[rand.Intn(len(states))]
	zip := randomZipCode()

	return fmt.Sprintf("%d %s, %s, %s %s", streetNumber, streetName, city, state, zip)
}
