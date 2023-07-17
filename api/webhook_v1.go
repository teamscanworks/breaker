package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// Typed integer representing operations to apply against the circuit breaker module
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

// A response returned from all webhook calls
type Response struct {
	// response message, if no errors set to "ok" otherwise includes the error
	Message string
	Urls    []string
	// transaction hash of any transaction(s) which were sent, if webhook failed
	// this is an empty string
	TxHash string
	// The operation that was applied to the circuit breaker
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
	var response Response
	if payload.Operation == MODE_TRIP {
		if tx, err := api.breakerClient.TripCircuitBreaker(r.Context(), payload.Urls); err != nil {
			api.logger.Error("failed to trip circuit breaker", zap.Error(err))
			response = Response{
				Message:   fmt.Sprintf("failed to trip circuit breaker %s", err),
				Urls:      payload.Urls,
				Operation: payload.Operation,
			}
		} else {
			response = Response{
				Message:   "ok",
				Urls:      payload.Urls,
				Operation: payload.Operation,
				TxHash:    tx,
			}
			api.logger.Info("tripped circuit", zap.Any("urls", payload.Urls))
		}
	} else if payload.Operation == MODE_RESET {
		if tx, err := api.breakerClient.ResetCircuitBreaker(r.Context(), payload.Urls); err != nil {
			api.logger.Error("failed to trip circuit breaker", zap.Error(err))
			response = Response{
				Message:   fmt.Sprintf("failed to reset circuit breaker %s", err),
				Urls:      payload.Urls,
				Operation: payload.Operation,
			}
		} else {
			response = Response{
				Message:   "ok",
				Urls:      payload.Urls,
				Operation: payload.Operation,
				TxHash:    tx,
			}
			api.logger.Info("reset circuit", zap.Any("urls", payload.Urls))
		}
	} else {
		response = Response{
			Message: fmt.Sprintf("unsupported mode %v", payload.Operation),
		}
	}
	rBytes, err := json.Marshal(&response)
	if err != nil {
		api.logger.Error("failed to serialize response", zap.Error(err))
		http.Error(w, "failed to serialize response", http.StatusInternalServerError)
	}
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(rBytes))
}
