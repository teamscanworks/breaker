package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

type API struct {
	router chi.Router
	// todo: add basic cache for pushed metrics
}

func NewAPI(ctx context.Context, log *zap.Logger, listenAddress string, username string, password string) (*API, error) {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	// todo: replace with JWT
	r.Use(middleware.BasicAuth("breaker", map[string]string{"admin": "admin", username: password}))
	api := API{router: r}
	api.router.Post("/push/metrics", api.HandlePushMetrics)
	return &api, nil
}

func (api *API) HandlePushMetrics(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("todo"))
}
