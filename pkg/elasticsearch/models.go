package elasticsearch

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	db "github.com/hankimmy/PtmrBackend/pkg/db/sqlc"
	"github.com/hankimmy/PtmrBackend/pkg/google"
	"github.com/hankimmy/PtmrBackend/pkg/util"
)

type Job struct {
	ID                 string          `json:"id"`
	EmployerID         int64           `json:"employer_id"`
	HiringOrganization string          `json:"hiring_organization"`
	Title              string          `json:"title"`
	Industry           string          `json:"industry"`
	JobLocation        string          `json:"job_location"`
	DatePosted         string          `json:"date_posted"`
	Description        string          `json:"description"`
	EmploymentType     string          `json:"employment_type"`
	Wage               float32         `json:"wage"`
	Tips               float32         `json:"tips"`
	DestinationURL     string          `json:"destination_URL"`
	IsUserCreated      bool            `json:"user_created"`
	JobApplication     json.RawMessage `json:"job_application"`
	// Google Business Data Related
	PlaceID          string              `json:"place_id"`
	DisplayName      string              `json:"display_name"`
	PhoneNumber      string              `json:"phone_number"`
	BusinessType     []string            `json:"business_types"`
	FormattedAddress string              `json:"formatted_address"`
	PreciseLocation  GeoPoint            `json:"precise_location"`
	Photos           []google.Photo      `json:"photos"`
	Rating           float32             `json:"rating"`
	PriceLevel       string              `json:"price_level"`
	OpeningHours     google.OpeningHours `json:"opening_hours"`
	WebsiteURI       string              `json:"website_uri"`
	GoogleMapsURI    string              `json:"google_maps_uri"`
}

type GeoPoint struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type esCandidate struct {
	ID                 int64               `json:"id"`
	Username           string              `json:"username"`
	FullName           string              `json:"full_name"`
	PhoneNumber        string              `json:"phone_number"`
	Education          db.Education        `json:"education"`
	Location           string              `json:"location"`
	SkillSet           []string            `json:"skill_set"`
	IndustryOfInterest string              `json:"industry_of_interest"`
	JobPreference      db.JobPreference    `json:"job_preference"`
	TimeAvailability   []util.Availability `json:"time_availability"`
	AccountVerified    bool                `json:"account_verified"`
	ResumeFile         string              `json:"resume_file"`
	ProfilePhoto       string              `json:"profile_photo"`
	Description        string              `json:"description"`
	CreatedAt          time.Time           `json:"created_at"`
}

type Candidate struct {
	UserUid            string        `json:"user_uid"`
	FullName           string        `json:"full_name"`
	Email              string        `json:"email"`
	PhoneNumber        string        `json:"phone_number"`
	Education          Education     `json:"education"`
	Location           string        `json:"location"`
	SkillSet           []string      `json:"skill_set"`
	Certificates       []string      `json:"certificates"`
	IndustryOfInterest string        `json:"industry_of_interest"`
	JobPreference      JobPreference `json:"job_preference"`
	TimeAvailability   []byte        `json:"time_availability"`
	ResumeFile         string        `json:"resume_file"`
	ProfilePhoto       string        `json:"profile_photo"`
	Description        string        `json:"description"`
	CreatedAt          time.Time     `json:"created_at"`
}

type PastExperience struct {
	ID          string    `json:"id"`
	Industry    string    `json:"industry"`
	Employer    string    `json:"employer"`
	JobTitle    string    `json:"job_title"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Present     bool      `json:"present"`
	Description string    `json:"description"`
}

type Education string

const (
	EducationHighSchoolDiploma Education = "high school diploma"
	EducationAssociate         Education = "associate"
	EducationBachelor          Education = "bachelor"
	EducationMaster            Education = "master"
	EducationDoctoral          Education = "doctoral"
)

type JobPreference string

const (
	JobPreferenceInPerson JobPreference = "in person"
	JobPreferenceRemote   JobPreference = "remote"
	JobPreferenceOpen     JobPreference = "open"
	JobPreferenceHybrid   JobPreference = "hybrid"
)

func RandomJob(employerID int64) Job {
	title := util.RandomString(5)
	applicationQuestions := []map[string]interface{}{
		{
			"id":         "q1",
			"type":       "text",
			"question":   "What is your full name?",
			"isRequired": true,
			"order":      1,
		},
		{
			"id":         "q3",
			"type":       "multiple-choice",
			"question":   "What is your highest level of education?",
			"options":    []string{"High School", "Associate Degree", "Bachelor's Degree", "Master's Degree", "Doctorate"},
			"isRequired": true,
			"order":      2,
		},
		{
			"id":         "q4",
			"type":       "yes-no",
			"question":   "Do you have a valid driver's license?",
			"isRequired": false,
			"order":      3,
		},
	}

	// Marshal the application questions into JSON
	jobApplication, err := json.Marshal(applicationQuestions)
	if err != nil {
		panic("failed to marshal job application questions: " + err.Error())
	}
	return Job{
		ID:                 fmt.Sprintf("%d_%s", employerID, title),
		EmployerID:         employerID,
		HiringOrganization: "Part Timer",
		Title:              title,
		Industry:           "Restaurant",
		JobLocation:        util.RandomUSAddress(),
		DatePosted:         time.Now().Format("2006-January-02"),
		Description:        util.RandomString(30),
		EmploymentType:     util.RandomString(5),
		Wage:               rand.Float32(),
		Tips:               rand.Float32(),
		DestinationURL:     util.RandomString(5),
		IsUserCreated:      true,
		JobApplication:     jobApplication,
		PlaceID:            "ChIJJS3mqONZwokR9KlP3H_7MNg",
		DisplayName:        "CHILI",
		PhoneNumber:        "(646) 882-0666",
		BusinessType:       []string{"chinese_restaurant", "establishment"},
		FormattedAddress:   "13 E 37th St, New York, NY 10016, USA",
		PreciseLocation:    GeoPoint{Lat: 40.7501259, Lon: -73.9820676},
		Photos: []google.Photo{
			{
				Name:     "places/ChIJJS3mqONZwokR9KlP3H_7MNg/photos/AelY_Ctep0GhoWSyGtUQcVSMWcruhQSVd5bg9XszHdMsNHAg3lejVc7VzzQ93nYqSkk7LRWqGHbh3AcCFeg8zVLXIa0KT7ZoH42RHBH26qvmJ8OJWIAbVjtCclz4jmjWESvvvS-aCiN-nTCV0CDK7oxuUbY_799xLJWhw5OA",
				WidthPx:  100,
				HeightPx: 100,
			},
		},
		Rating:        rand.Float32() + 4,
		PriceLevel:    "PRICE_LEVEL_MODERATE",
		OpeningHours:  google.OpeningHours{},
		WebsiteURI:    "https://www.chilinyc.com/",
		GoogleMapsURI: "https://www.google.com/maps/place/CHILI/@40.7501259,-73.9820676,15z/data=!4m2!3m1!1s0x0:0xd830fb7fdc4fa9f4?sa=X&ved=1t:2428&ictx=111",
	}
}

func RandomEducation() Education {
	educations := []Education{
		EducationHighSchoolDiploma,
		EducationAssociate,
		EducationBachelor,
		EducationMaster,
		EducationDoctoral,
	}
	return educations[util.RandomInt(0, 4)]
}

func RandomJobPref() JobPreference {
	jobPreferences := []JobPreference{
		JobPreferenceInPerson,
		JobPreferenceOpen,
		JobPreferenceHybrid,
		JobPreferenceRemote,
	}
	return jobPreferences[util.RandomInt(0, 3)]
}
