package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"go.uber.org/zap"
)

type Payload struct {
	Message string
	Urls    []string
}

func (api *API) HandleWebookV1(w http.ResponseWriter, r *http.Request) {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var payload Payload
	if err = json.Unmarshal(data, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	api.logger.Info("tripping circuit", zap.String("message", payload.Message), zap.Any("urls", payload.Urls), zap.Bool("dry.run", api.dryRun))
	if api.dryRun {
		w.Write([]byte("dry run, skipping transaction invocation"))
		return
	}
	if api.breakerClient == nil {
		http.Error(w, "no initialized breaker client", http.StatusInternalServerError)
		return
	}
	if err := api.breakerClient.TripCircuitBreaker(r.Context(), payload.Urls); err != nil {
		api.logger.Error("failed to trip circuit breaker", zap.Error(err))
	} else {
		api.logger.Info("tripped circuit", zap.Any("urls", payload.Urls))
	}
}
