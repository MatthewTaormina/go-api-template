package handlers

import (
	"net/http"
	"time"

	"github.com/user/go-api-template/internal/respond"
)

// healthResponse is the shape returned by GET /health.
type healthResponse struct {
	Status string    `json:"status"`
	Time   time.Time `json:"time"`
}

// Health handles GET /health.
// It is used by load balancers, reverse proxies, and container orchestrators
// to determine whether the server is alive and ready to receive traffic.
// Always returns 200 as long as the process is running.
func Health(w http.ResponseWriter, r *http.Request) {
	respond.JSON(w, http.StatusOK, healthResponse{
		Status: "ok",
		Time:   time.Now().UTC(),
	})
}
