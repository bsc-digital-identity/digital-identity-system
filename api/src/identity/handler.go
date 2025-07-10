package identity

import (
	"encoding/json"
	"gorm.io/gorm"
	"net/http"
)

type createRequest struct {
	IdentityName string `json:"identity_name"`
}

type createResponse struct {
	IdentityId   string `json:"identity_id"`
	IdentityName string `json:"identity_name"`
}

func MakeCreateHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req createRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IdentityName == "" {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		identity, err := CreateIdentity(db, req.IdentityName)
		if err != nil {
			http.Error(w, "Could not create identity: "+err.Error(), http.StatusInternalServerError)
			return
		}
		resp := createResponse{
			IdentityId:   identity.IdentityId,
			IdentityName: identity.IdentityName,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
