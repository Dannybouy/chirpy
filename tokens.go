package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dannybouy/chirpy/internal/auth"
	"github.com/dannybouy/chirpy/internal/database"
)

var errInvalidRefreshToken = errors.New("invalid refresh token")

type TokenResponse struct {
	AccessToken string `json:"token"`
}

func (apiCfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := apiCfg.validateRefreshToken(r)
	if errors.Is(err, errInvalidRefreshToken) {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't look up refresh token", err)
		return
	}

	accessToken, err := auth.MakeJWT(refreshToken.UserID, apiCfg.jwtSecret, time.Hour)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create access token", err)
		return
	}

	respondWithJSON(w, http.StatusOK, TokenResponse{
		AccessToken: accessToken,
	})
}

func (apiCfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Invalid refresh token", err)
		return
	}

	err = apiCfg.db.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't revoke refresh token", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// validateRefreshToken returns an error wrapping errInvalidRefreshToken when the
// client should get a 401; any other error is a server-side failure.
func (apiCfg *apiConfig) validateRefreshToken(r *http.Request) (database.RefreshToken, error) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		return database.RefreshToken{}, fmt.Errorf("%w: %v", errInvalidRefreshToken, err)
	}

	verifiedRefreshToken, err := apiCfg.db.GetRefreshToken(r.Context(), refreshToken)
	if errors.Is(err, sql.ErrNoRows) {
		return database.RefreshToken{}, fmt.Errorf("%w: not found", errInvalidRefreshToken)
	}
	if err != nil {
		return database.RefreshToken{}, err
	}

	if verifiedRefreshToken.ExpiresAt.Before(time.Now().UTC()) {
		return database.RefreshToken{}, fmt.Errorf("%w: expired", errInvalidRefreshToken)
	}

	if verifiedRefreshToken.RevokedAt.Valid {
		return database.RefreshToken{}, fmt.Errorf("%w: revoked", errInvalidRefreshToken)
	}

	return verifiedRefreshToken, nil
}
