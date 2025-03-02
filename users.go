package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/Chase-Outman/GitLab/internal/auth"
	"github.com/Chase-Outman/GitLab/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUpdateUsers(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	userP := userParams{}

	err := decoder.Decode(&userP)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed decoding parameters")
		return
	}

	hashedPassword, err := auth.HashPassword(userP.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = cfg.db.UpdateUser(r.Context(), database.UpdateUserParams{
		Email:          userP.Email,
		HashedPassword: hashedPassword,
		ID:             r.Context().Value("userID").(uuid.UUID),
	})
	if err != nil {
		log.Printf("Error updating user: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	type userReturnVals struct {
		Email string `json:"email"`
	}

	resp := userReturnVals{
		Email: userP.Email,
	}

	respondWithJSON(w, http.StatusOK, resp)
}

func (cfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type userParams struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	userP := userParams{}

	err := decoder.Decode(&userP)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(userP.Password)
	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	user, err := cfg.db.CreateUser(r.Context(), database.CreateUserParams{
		Email:          userP.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		log.Printf("Error create new user: %s", err)
		w.WriteHeader(500)
		return
	}

	type userReturnVals struct {
		Id          uuid.UUID `json:"id"`
		Created_at  time.Time `json:"created_at"`
		Updated_at  time.Time `json:"updated_at"`
		Email       string    `json:"email"`
		IsChirpyRed bool      `json:"is_chirpy_red"`
	}

	resp := userReturnVals{
		Id:          user.ID,
		Created_at:  user.CreatedAt,
		Updated_at:  user.UpdatedAt,
		Email:       user.Email,
		IsChirpyRed: user.IsChirpyRed,
	}

	respondWithJSON(w, 201, resp)

}
