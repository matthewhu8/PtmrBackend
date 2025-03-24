package db

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/hankimmy/PtmrBackend/pkg/util"
)

func RandomUser(role Role) (user User, password string) {
	password = util.RandomString(6)
	hashedPassword, _ := util.HashPassword(password)

	user = User{
		Username:       util.RandomString(6),
		Email:          util.RandomEmail(),
		HashedPassword: hashedPassword,
		Role:           role,
	}
	return user, hashedPassword
}

func RandomCandidate(username string) Candidate {
	return Candidate{
		ID:                 util.RandomInt(1, 1000),
		Username:           username,
		FullName:           util.RandomString(10),
		PhoneNumber:        util.RandomPhoneNumber(),
		Education:          RandomEducation(),
		Location:           util.RandomString(20),
		SkillSet:           []string{util.RandomString(5), util.RandomString(5)},
		IndustryOfInterest: util.RandomString(10),
		JobPreference:      RandomJobPref(),
		TimeAvailability:   util.RandomAvailability(),
		AccountVerified:    util.RandomBool(),
		ResumeFile:         util.RandomString(10),
		ProfilePhoto:       util.RandomString(10),
		Description:        util.RandomString(50),
	}
}

func RandomEmployer(username string) Employer {
	return Employer{
		ID:                  util.RandomInt(1, 1000),
		Username:            username,
		BusinessName:        util.RandomString(10),
		BusinessEmail:       util.RandomEmail(),
		BusinessPhone:       util.RandomPhoneNumber(),
		Location:            util.RandomString(20),
		Industry:            util.RandomString(10),
		ProfilePhoto:        util.RandomString(10),
		BusinessDescription: util.RandomString(50),
	}
}

func RandomCandidateApplication(employerID int64) CandidateApplication {
	cID := util.RandomInt(0, 100)
	return CandidateApplication{
		CandidateID:        cID,
		EmployerID:         employerID,
		ElasticsearchDocID: fmt.Sprintf("%d_%d", cID, employerID),
		ApplicationStatus:  ApplicationStatusPending,
		CreatedAt:          time.Time{},
	}
}

func RandomEmployerApplication(candidateID int64) EmployerApplication {
	eID := util.RandomInt(0, 100)
	return EmployerApplication{
		CandidateID:       candidateID,
		EmployerID:        eID,
		Message:           util.RandomString(10),
		ApplicationStatus: ApplicationStatusPending,
		CreatedAt:         time.Time{},
	}
}

func RandomJobPref() JobPreference {
	jobPreferences := []JobPreference{
		JobPreferenceInperson,
		JobPreferenceOpen,
		JobPreferenceHybrid,
		JobPreferenceRemote,
	}
	return jobPreferences[util.RandomInt(0, 3)]
}

func RandomEducation() Education {
	educations := []Education{
		EducationHighschooldiploma,
		EducationAssociate,
		EducationBachelor,
		EducationMaster,
		EducationDoctoral,
	}
	return educations[util.RandomInt(0, 4)]
}

func RandomDateRange(year int) (pgtype.Date, pgtype.Date) {
	startOfYear := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	endOfYear := time.Date(year, 12, 31, 0, 0, 0, 0, time.UTC)
	randomStartDate := randomDate(startOfYear, endOfYear).UTC().Truncate(24 * time.Hour)
	randomEndDate := randomDate(randomStartDate, endOfYear).UTC().Truncate(24 * time.Hour)
	startDate := pgtype.Date{
		Time:             randomStartDate,
		InfinityModifier: 0, // 0 represents a normal date (no infinity modifier)
		Valid:            true,
	}
	endDate := pgtype.Date{
		Time:             randomEndDate,
		InfinityModifier: 0,
		Valid:            true,
	}
	return startDate, endDate
}

func randomDate(min, max time.Time) time.Time {
	delta := max.Unix() - min.Unix()
	seconds := rand.Int63n(delta)
	return time.Unix(min.Unix()+seconds, 0)
}
