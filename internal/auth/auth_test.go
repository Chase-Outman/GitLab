package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGetBearerToken(t *testing.T) {
	userId1 := uuid.New()
	validToken, _ := MakeJWT(userId1, "secret", time.Hour)
	header := http.Header{}
	header.Set("Authorization", validToken)

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "Valid authorization",
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotToken, err := GetBearerToken(header)
			if (err != nil) != tc.wantErr {
				t.Errorf("GetBearerToken() error = %v wantErr %v", err, tc.wantErr)
			}
			if gotToken != validToken {
				t.Errorf("GetBearerToken() gotToken = %v want %v", gotToken, validToken)
			}
		})
	}
}

func TestCreateAndValidateJWT(t *testing.T) {
	userId1 := uuid.New()
	validToken, _ := MakeJWT(userId1, "secret", time.Hour)

	tests := []struct {
		name        string
		tokenString string
		tokenSecret string
		wantUserID  uuid.UUID
		wantErr     bool
	}{
		{
			name:        "Valid token",
			tokenString: validToken,
			tokenSecret: "secret",
			wantUserID:  userId1,
			wantErr:     false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotUserID, err := ValidateJWT(tc.tokenString, tc.tokenSecret)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateJWT() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if gotUserID != tc.wantUserID {
				t.Errorf("ValidateJWT() gotUserID = %v, want %v", gotUserID, tc.wantUserID)
			}
		})
	}
}
