package db

import (
	"context"
	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func createRandomEmployer(t *testing.T) Employer {
	user := createRandomUser(t)

	arg := CreateEmployerParams{
		Username:            user.Username,
		BusinessName:        util.RandomString(10),
		BusinessEmail:       util.RandomEmail(),
		BusinessPhone:       util.RandomPhoneNumber(),
		Location:            util.RandomString(20),
		Industry:            util.RandomString(10),
		ProfilePhoto:        util.RandomString(10),
		BusinessDescription: util.RandomString(50),
	}

	employer, err := testStore.CreateEmployer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, employer)

	require.Equal(t, arg.Username, employer.Username)
	require.Equal(t, arg.BusinessName, employer.BusinessName)
	require.Equal(t, arg.BusinessEmail, employer.BusinessEmail)
	require.Equal(t, arg.BusinessPhone, employer.BusinessPhone)
	require.Equal(t, arg.Location, employer.Location)
	require.Equal(t, arg.Industry, employer.Industry)
	require.Equal(t, arg.ProfilePhoto, employer.ProfilePhoto)
	require.Equal(t, arg.BusinessDescription, employer.BusinessDescription)

	require.NotZero(t, employer.ID)
	require.NotZero(t, employer.CreatedAt)

	return employer
}

func TestCreateEmployer(t *testing.T) {
	createRandomEmployer(t)
}

func TestGetEmployer(t *testing.T) {
	employer1 := createRandomEmployer(t)
	employer2, err := testStore.GetEmployer(context.Background(), employer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, employer2)

	require.Equal(t, employer1.ID, employer2.ID)
	require.Equal(t, employer1.Username, employer2.Username)
	require.Equal(t, employer1.BusinessName, employer2.BusinessName)
	require.Equal(t, employer1.BusinessEmail, employer2.BusinessEmail)
	require.Equal(t, employer1.BusinessPhone, employer2.BusinessPhone)
	require.Equal(t, employer1.Location, employer2.Location)
	require.Equal(t, employer1.Industry, employer2.Industry)
	require.Equal(t, employer1.ProfilePhoto, employer2.ProfilePhoto)
	require.Equal(t, employer1.BusinessDescription, employer2.BusinessDescription)
	require.WithinDuration(t, employer1.CreatedAt, employer2.CreatedAt, time.Second)
}

func TestUpdateEmployerBusinessEmail(t *testing.T) {
	employer1 := createRandomEmployer(t)

	email := util.RandomEmail()
	arg := UpdateEmployerParams{
		ID: employer1.ID,
		BusinessEmail: pgtype.Text{
			String: email,
			Valid:  true,
		},
	}

	updatedEmployer, err := testStore.UpdateEmployer(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, employer1.ID, updatedEmployer.ID)
	require.Equal(t, employer1.BusinessName, updatedEmployer.BusinessName)
	require.Equal(t, email, updatedEmployer.BusinessEmail)
	require.Equal(t, employer1.BusinessPhone, updatedEmployer.BusinessPhone)
	require.Equal(t, employer1.Location, updatedEmployer.Location)
	require.Equal(t, employer1.Industry, updatedEmployer.Industry)
	require.Equal(t, employer1.ProfilePhoto, updatedEmployer.ProfilePhoto)
	require.Equal(t, employer1.BusinessDescription, updatedEmployer.BusinessDescription)
	require.WithinDuration(t, employer1.CreatedAt, updatedEmployer.CreatedAt, time.Second)
}

func TestUpdateEmployerIndustryAndDescription(t *testing.T) {
	employer1 := createRandomEmployer(t)

	updatedField := util.RandomString(10)
	arg := UpdateEmployerParams{
		ID: employer1.ID,
		Industry: pgtype.Text{
			String: updatedField,
			Valid:  true,
		},
		BusinessDescription: pgtype.Text{
			String: updatedField,
			Valid:  true,
		},
	}

	updatedEmployer, err := testStore.UpdateEmployer(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, employer1.ID, updatedEmployer.ID)
	require.Equal(t, employer1.BusinessName, updatedEmployer.BusinessName)
	require.Equal(t, employer1.BusinessEmail, updatedEmployer.BusinessEmail)
	require.Equal(t, employer1.BusinessPhone, updatedEmployer.BusinessPhone)
	require.Equal(t, employer1.Location, updatedEmployer.Location)
	require.Equal(t, updatedField, updatedEmployer.Industry)
	require.Equal(t, employer1.ProfilePhoto, updatedEmployer.ProfilePhoto)
	require.Equal(t, updatedField, updatedEmployer.BusinessDescription)
	require.WithinDuration(t, employer1.CreatedAt, updatedEmployer.CreatedAt, time.Second)
}

func TestUpdateEmployerAllFields(t *testing.T) {
	employer1 := createRandomEmployer(t)

	businessName := util.RandomString(10)
	businessEmail := util.RandomEmail()
	businessPhone := util.RandomPhoneNumber()
	location := util.RandomString(10)
	arg := UpdateEmployerParams{
		ID: employer1.ID,
		BusinessName: pgtype.Text{
			String: businessName,
			Valid:  true,
		},
		BusinessEmail: pgtype.Text{
			String: businessEmail,
			Valid:  true,
		},
		BusinessPhone: pgtype.Text{
			String: businessPhone,
			Valid:  true,
		},
		Location: pgtype.Text{
			String: location,
			Valid:  true,
		},
		Industry: pgtype.Text{
			String: location,
			Valid:  true,
		},
		ProfilePhoto: pgtype.Text{
			String: location,
			Valid:  true,
		},
		BusinessDescription: pgtype.Text{
			String: location,
			Valid:  true,
		},
	}

	updatedUser, err := testStore.UpdateEmployer(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, employer1.ID, updatedUser.ID)
	require.Equal(t, employer1.Username, updatedUser.Username)
	require.Equal(t, businessName, updatedUser.BusinessName)
	require.Equal(t, businessEmail, updatedUser.BusinessEmail)
	require.Equal(t, businessPhone, updatedUser.BusinessPhone)
	require.Equal(t, location, updatedUser.Location)
	require.Equal(t, location, updatedUser.Industry)
	require.Equal(t, location, updatedUser.ProfilePhoto)
	require.Equal(t, location, updatedUser.BusinessDescription)
	require.WithinDuration(t, employer1.CreatedAt, updatedUser.CreatedAt, time.Second)
}

func TestDeleteEmployer(t *testing.T) {
	employer1 := createRandomEmployer(t)
	err := testStore.DeleteEmployer(context.Background(), employer1.ID)
	require.NoError(t, err)

	employer2, err := testStore.GetEmployer(context.Background(), employer1.ID)
	require.Error(t, err)
	require.EqualError(t, err, ErrRecordNotFound.Error())
	require.Empty(t, employer2)
}

func TestListEmployers(t *testing.T) {
	var lastEmployer Employer
	for i := 0; i < 10; i++ {
		lastEmployer = createRandomEmployer(t)
	}

	arg := ListEmployersParams{
		Username: lastEmployer.Username,
		Limit:    5,
		Offset:   0,
	}

	employers, err := testStore.ListEmployers(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, employers)

	for _, employer := range employers {
		require.NotEmpty(t, employer)
	}
}

func TestAddJobListing(t *testing.T) {
	employer := createRandomEmployer(t)
	arg := AddJobListingParams{
		ID:    employer.ID,
		JobID: "New Job",
	}
	err1 := testStore.AddJobListing(context.Background(), arg)
	require.NoError(t, err1)

	res, err2 := testStore.GetEmployer(context.Background(), employer.ID)
	require.NoError(t, err2)
	require.Equal(t, res.JobListings[0], "New Job")
}
