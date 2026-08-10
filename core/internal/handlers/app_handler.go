package handlers

import (
	"encoding/json"
	"net/http"
	"starvault/core/internal/dtos"
	"starvault/core/internal/services"
)

type AppHandler struct {
	AppService *services.AppService
}

func (h *AppHandler) CreateApp(w http.ResponseWriter, r *http.Request) {
	var req dtos.CreateAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	res, err := h.AppService.CreateApp(req)
	if err != nil {
		if err.Error() == "name is required" || err.Error() == "name is too long" {
			sendError(w, http.StatusBadRequest, err.Error())
		} else {
			sendError(w, http.StatusInternalServerError, "Internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}
