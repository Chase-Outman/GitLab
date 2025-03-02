package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Chase-Outman/GitLab/internal/auth"
	"github.com/Chase-Outman/GitLab/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerLogin(w http.ResponseWriter, r *http.Request) {
	type loginParams struct {
		Password         string `json:"password"`
		Email            string `json:"email"`
		ExpiresInSeconds *int   `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	login := loginParams{}
	err := decoder.Decode(&login)
	if err != nil {
		return
	}
	var expiresIn int
	if login.ExpiresInSeconds == nil {
		expiresIn = 3600
	} else {
		expiresIn = *login.ExpiresInSeconds
		if expiresIn > 3600 {
			expiresIn = 3600
		}
	}

	user, err := cfg.db.GetUserByEmail(r.Context(), login.Email)
	if err != nil {
		log.Printf("Error getting user by email: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = auth.CheckPasswordHash(login.Password, user.HashedPassword)
	if err != nil {
		log.Printf("Error checking password: %s", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Duration(expiresIn)*time.Second)
	if err != nil {
		log.Printf("Error making JWT: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		log.Printf("Error making refresh token: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	hoursIn60Days := 1440
	_, _ = cfg.db.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(time.Duration(hoursIn60Days) * time.Hour),
	})

	type userReturnVals struct {
		Id           uuid.UUID `json:"id"`
		Created_at   time.Time `json:"created_at"`
		Updated_at   time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
		IsChipryRed  bool      `json:"is_chirpy_red"`
	}

	respondWithJSON(w, http.StatusOK, userReturnVals{
		Id:           user.ID,
		Created_at:   user.CreatedAt,
		Updated_at:   user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
		IsChipryRed:  user.IsChirpyRed,
	})

}

func (cfg *apiConfig) middlewareAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userUUID, err := auth.ValidateJWT(token, cfg.secret)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userID", userUUID)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	refreshToken, err := cfg.db.GetRefreshToken(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if refreshToken.RevokedAt.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	jwt, err := auth.MakeJWT(refreshToken.UserID, cfg.secret, time.Duration(3600)*time.Second)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	type returnVal struct {
		Token string `json:"token"`
	}

	respondWithJSON(w, http.StatusOK, returnVal{
		Token: jwt,
	})

}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	err = cfg.db.RevokeRefreshToken(r.Context(), token)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
