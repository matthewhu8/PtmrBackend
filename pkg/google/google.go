package google

type GAPI interface {
	GetLatLon(address string) (float64, float64, error)
	GetPlaceIDOfAddress(address string) (string, error)
	GetPlaceID(query string) (string, error)
	GetPlaceDetails(placeID string) (*PlaceDetailsResponse, error)
}

type Service struct{}

func NewGoogleService() *Service {
	return &Service{}
}
