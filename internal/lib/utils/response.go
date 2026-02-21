package utils

import (
	"encoding/json"
	"net/http"

	"github.com/AdityaTote/wallet-service/internal/models"
)


func JSONWriter(w http.ResponseWriter, status int, data models.JSONResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(data)
}