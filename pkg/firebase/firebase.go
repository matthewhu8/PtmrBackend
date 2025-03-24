package firebase

import (
	"context"
	"fmt"

	"firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

const redirectUrl = "http://localhost:9000/setup"

type AuthClientFirebase interface {
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
	SetUserRole(ctx context.Context, uid string, role string) error
	GenerateEmailVerificationLink(ctx context.Context, email, role string) (string, error)
	GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error)
	IsEmailVerified(ctx context.Context, email string) (bool, error)
}

type AuthClient struct {
	authClient *auth.Client
}

func NewAuthClient(credentialsFile string) (*AuthClient, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	ctx := context.Background()
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}

	return &AuthClient{authClient: client}, nil
}

func (f *AuthClient) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return f.authClient.VerifyIDToken(ctx, idToken)
}

func (f *AuthClient) SetUserRole(ctx context.Context, uid string, role string) error {
	claims := map[string]interface{}{
		"role": role,
	}
	params := (&auth.UserToUpdate{}).CustomClaims(claims)
	_, err := f.authClient.UpdateUser(ctx, uid, params)
	return err
}

func (f *AuthClient) GenerateEmailVerificationLink(ctx context.Context, email, role string) (string, error) {
	var redirectUrl string
	if role == "candidate" {
		redirectUrl = "http://localhost:9000/candidate-setup"
	} else if role == "employer" {
		redirectUrl = "http://localhost:9000/employer-setup"
	} else {
		return "", fmt.Errorf("invalid role")
	}

	actionCodeSettings := &auth.ActionCodeSettings{
		URL:             redirectUrl,
		HandleCodeInApp: false,
	}

	return f.authClient.EmailVerificationLinkWithSettings(ctx, email, actionCodeSettings)
}

func (f *AuthClient) GetUserByEmail(ctx context.Context, email string) (*auth.UserRecord, error) {
	return f.authClient.GetUserByEmail(ctx, email)
}

func (f *AuthClient) IsEmailVerified(ctx context.Context, email string) (bool, error) {
	userRecord, err := f.GetUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	return userRecord.EmailVerified, nil
}
