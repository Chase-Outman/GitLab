package main

import (
	"net/http"

	"github.com/Chase-Outman/GitLab/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	if chirpID == "" {
		http.NotFound(w, r)
		return
	}

	chirpUuid, err := uuid.Parse(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to parse chirpID to uuid")
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "failed to authorize user")
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		respondWithError(w, http.StatusForbidden, "failed to validate JWT")
		return
	}

	chirp, err := cfg.db.GetChirpById(r.Context(), chirpUuid)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "failed to find chirp with id")
		return
	}

	if chirp.UserID != userID {
		respondWithError(w, http.StatusForbidden, "user not allowed to delete chirp")
		return
	}

	err = cfg.db.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to delete chirp from database")
		return
	}

	respondWithJSON(w, http.StatusNoContent, nil)

}
