package db

import (
	"context"
	"testing"
	"time"

	"github.com/hankimmy/PtmrBackend/pkg/util"
	"github.com/stretchr/testify/require"
)

// createRandomVerifyEmail creates a random verify email entry for testing.
func createRandomVerifyEmail(t *testing.T) VerifyEmail {
	user := createRandomUser(t)

	arg := CreateVerifyEmailParams{
		Username:   user.Username,
		Email:      util.RandomEmail(),
		SecretCode: util.RandomString(32),
	}

	verifyEmail, err := testStore.CreateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, verifyEmail)

	require.Equal(t, arg.Username, verifyEmail.Username)
	require.Equal(t, arg.Email, verifyEmail.Email)
	require.Equal(t, arg.SecretCode, verifyEmail.SecretCode)
	require.False(t, verifyEmail.IsUsed)
	require.NotZero(t, verifyEmail.CreatedAt)
	require.NotZero(t, verifyEmail.ExpiredAt)

	return verifyEmail
}

func TestCreateVerifyEmail(t *testing.T) {
	createRandomVerifyEmail(t)
}

func TestUpdateVerifyEmail(t *testing.T) {
	t.Skip()
	verifyEmail := createRandomVerifyEmail(t)

	arg := UpdateVerifyEmailParams{
		ID:         verifyEmail.ID,
		SecretCode: verifyEmail.SecretCode,
	}

	updatedVerifyEmail, err := testStore.UpdateVerifyEmail(context.Background(), arg)
	require.NoError(t, err)

	require.Equal(t, verifyEmail.ID, updatedVerifyEmail.ID)
	require.Equal(t, verifyEmail.Username, updatedVerifyEmail.Username)
	require.Equal(t, verifyEmail.Email, updatedVerifyEmail.Email)
	require.Equal(t, verifyEmail.SecretCode, updatedVerifyEmail.SecretCode)
	require.True(t, updatedVerifyEmail.IsUsed)
	require.WithinDuration(t, verifyEmail.CreatedAt, updatedVerifyEmail.CreatedAt, time.Second)
	require.WithinDuration(t, verifyEmail.ExpiredAt, updatedVerifyEmail.ExpiredAt, time.Second)
}
