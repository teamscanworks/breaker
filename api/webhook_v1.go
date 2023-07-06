package api

import (
	"encoding/json"
	"io"
	"net/http"

	"go.uber.org/zap"
)

type Mode int

const (
	MODE_TRIP Mode = iota // 0
	MODE_RESET
)

// The payload that can be sent through the v1 webhook API
type PayloadV1 struct {
	// message that is logged, should be the reason for tripping a circuit
	Message string
	// module request urls that we want to apply some action to
	Urls []string
	// the operation being applied against the circuit breaker module
	Operation Mode
}

// Function which handles the webhook api call for V1 payloads, and consists of
// deserializing a PayloadV1 message type. You may specific one of two modes, either
// tripping or resetting a circuit for a list of module request urls.
func (api *API) HandleWebookV1(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	var payload PayloadV1
	if err = json.Unmarshal(data, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var msg string
	if payload.Operation == MODE_TRIP {
		msg = "tripping circuit"
	} else if payload.Operation == MODE_RESET {
		msg = "resetting circuit"
	} else {
		http.Error(w, "unsupported mode", http.StatusBadRequest)
		return
	}
	api.logger.Info(msg, zap.String("message", payload.Message), zap.Any("urls", payload.Urls), zap.Bool("dry.run", api.dryRun))
	if api.dryRun {
		w.Write([]byte("dry run, skipping transaction invocation"))
		return
	}
	if api.breakerClient == nil {
		http.Error(w, "no initialized breaker client", http.StatusInternalServerError)
		return
	}
	if payload.Operation == MODE_TRIP {
		if err := api.breakerClient.TripCircuitBreaker(r.Context(), payload.Urls); err != nil {
			api.logger.Error("failed to trip circuit breaker", zap.Error(err))
		} else {
			api.logger.Info("tripped circuit", zap.Any("urls", payload.Urls))
		}
	} else if payload.Operation == MODE_RESET {
		if err := api.breakerClient.ResetCircuitBreaker(r.Context(), payload.Urls); err != nil {
			api.logger.Error("failed to trip circuit breaker", zap.Error(err))
		} else {
			api.logger.Info("tripped circuit", zap.Any("urls", payload.Urls))
		}
	} else {
		http.Error(w, "unsupported mode", http.StatusBadRequest)
		return
	}

}
