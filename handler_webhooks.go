package main

import (
	"encoding/json"
	"net/http"

	"github.com/Chase-Outman/GitLab/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerWebhooks(w http.ResponseWriter, r *http.Request) {
	type eventRequest struct {
		Event string `json:"event"`
		Data  struct {
			UserID string `json:"user_id"`
		} `json:"data"`
	}

	gotAPIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to get api key")
		return
	}

	if gotAPIKey != cfg.apiKey {
		respondWithError(w, http.StatusUnauthorized, "provided api key does not match needed api key")
		return
	}

	decoder := json.NewDecoder(r.Body)
	event := eventRequest{}
	err = decoder.Decode(&event)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to decode parameters")
		return
	}

	if event.Event != "user.upgraded" {
		respondWithError(w, http.StatusNoContent, "")
		return
	}

	userID, err := uuid.Parse(event.Data.UserID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse userID to UUID")
		return
	}

	_, err = cfg.db.UpgradeChirpyRed(r.Context(), userID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "failed to upgrade user to chirpy red")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)
}
