package main

import (
	"encoding/json"
	"log"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/Chase-Outman/GitLab/internal/database"
	"github.com/google/uuid"
)

const censor = "****"
const maxChirpLength = 140

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	UserID    uuid.UUID `json:"user_id"`
	Body      string    `json:"body"`
}

func (cfg *apiConfig) handlerGetChirpByID(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		http.NotFound(w, r)
		return
	}
	chirpUuid, err := uuid.Parse(chirpID)
	if err != nil {
		log.Printf("Error parse uuid: %s", err)
		w.WriteHeader(500)
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpUuid)
	if err != nil {
		log.Printf("Error getting chirp: %s", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	respondWithJSON(w, 200, Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.CreatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	})
}

func (cfg *apiConfig) handlersGetChirps(w http.ResponseWriter, r *http.Request) {
	sorted := r.URL.Query().Get("sort")
	if sorted == "" {
		sorted = "asc"
	}
	authorID := r.URL.Query().Get("author_id")
	chirps := []Chirp{}
	if authorID == "" {
		allChirps, err := cfg.db.GetChirps(r.Context())
		if err != nil {
			log.Printf("Error getting chirps in database: %s", err)
			w.WriteHeader(500)
			return
		}
		for _, dbChirp := range allChirps {
			chirps = append(chirps, Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.CreatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			})
		}
	} else {
		userID, err := uuid.Parse(authorID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to parse provided author id to UUID")
			return
		}
		authorChirps, err := cfg.db.GetChirpsByAuthor(r.Context(), userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "failed to get chirps with author id")
			return
		}
		for _, dbChirp := range authorChirps {
			chirps = append(chirps, Chirp{
				ID:        dbChirp.ID,
				CreatedAt: dbChirp.CreatedAt,
				UpdatedAt: dbChirp.CreatedAt,
				Body:      dbChirp.Body,
				UserID:    dbChirp.UserID,
			})
		}
	}
	if sorted == "desc" {
		sort.Slice(chirps, func(i, j int) bool { return chirps[i].CreatedAt.After(chirps[j].CreatedAt) })
	}

	respondWithJSON(w, http.StatusOK, chirps)

}

func (cfg *apiConfig) handlerChirps(w http.ResponseWriter, r *http.Request) {

	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}

	type parameters struct {
		Body string `json:"body"`
	}

	userID := r.Context().Value("userID").(uuid.UUID)

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	//validate chirp
	if len(params.Body) > maxChirpLength {
		respondWithError(w, 400, "Chirp is too long")

	} else {
		words := strings.Fields(params.Body)
		for _, word := range words {
			if slices.Contains(profaneWords, strings.ToLower(word)) {
				params.Body = strings.Replace(params.Body, word, censor, -1)
			}
		}
		chirp, err := cfg.db.CreateChirp(r.Context(), database.CreateChirpParams{
			Body:   params.Body,
			UserID: userID,
		})
		if err != nil {
			log.Printf("Error creating chirp in database: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		resp := Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		respondWithJSON(w, 201, resp)
	}

}
