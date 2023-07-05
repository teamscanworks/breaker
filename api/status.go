package api

import (
	"bytes"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func (api *API) ListDisabledCommands(w http.ResponseWriter, r *http.Request) {
	if api.breakerClient == nil {
		http.Error(w, "no initialized cosmos client", http.StatusInternalServerError)
		return
	}
	res, err := api.breakerClient.ListDisabledCommands(r.Context())
	if err != nil {
		api.logger.Error("failed to lsit disabled commands", zap.Error(err))
		http.Error(w, "failed to list disabled commands", http.StatusInternalServerError)
		return
	}
	data, err := res.Marshal()
	if err != nil {
		api.logger.Error("failed to marshal response", zap.Error(err))
		http.Error(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(data))
}

func (api *API) ListAccounts(w http.ResponseWriter, r *http.Request) {
	if api.breakerClient == nil {
		http.Error(w, "no initialized cosmos client", http.StatusInternalServerError)
		return
	}
	res, err := api.breakerClient.Accounts(r.Context())
	if err != nil {
		api.logger.Error("failed to list accounts", zap.Error(err))
		http.Error(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}
	data, err := res.Marshal()
	if err != nil {
		api.logger.Error("failed to marshal response", zap.Error(err))
		http.Error(w, "failed to list accounts", http.StatusInternalServerError)
		return
	}
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(data))
}
