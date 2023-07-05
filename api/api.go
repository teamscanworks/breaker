package api

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/teamscanworks/breaker/breakerclient"
	"go.uber.org/zap"
)

type API struct {
	ctx           context.Context
	cancel        context.CancelFunc
	router        chi.Router
	logger        *zap.Logger
	jwt           *JWT
	breakerClient *breakerclient.BreakerClient
	addr          string
	// if true, do not invoke any cosmos transactions
	dryRun bool
	// todo: add basic cache for pushed metrics
}

type ApiOpts struct {
	ListenAddress                string
	Password                     string
	IdentifierField              string
	TokenValidityDurationSeconds int64
	DryRun                       bool
}

func NewAPI(
	ctx context.Context,
	log *zap.Logger,
	jwt *JWT,
	opts ApiOpts,
) (*API, error) {
	ctx, cancel := context.WithCancel(ctx)
	logger := log.Named("breaker.api")
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(NewLoggerMiddleware(logger))
	api := API{ctx: ctx, cancel: cancel, router: r, jwt: NewJWT(opts.Password, opts.IdentifierField, opts.TokenValidityDurationSeconds), addr: opts.ListenAddress, logger: logger, dryRun: opts.DryRun}
	api.router.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(jwtauth.Verifier(api.jwt.tokenAuth))
			r.Use(api.jwt.Authenticator)
			r.Post("/webhook", api.HandleWebookV1)
		})
		r.Group(func(r chi.Router) {
			// status urls that can be used to return information
			r.Get("/status/listDisabledCommands", api.ListDisabledCommands)
			r.Get("/status/accounts", api.ListAccounts)
		})
	})
	return &api, nil
}

// sets the breakerClient field, needed for non dry-run webhook calls, as well as the status calls
func (api *API) WithBreakerClient(client *breakerclient.BreakerClient) {
	api.breakerClient = client
}

// cancels the api context, triggering a shutdown of the api router
// if `api.Server` has been called
func (api *API) Close() {
	api.cancel()
}

// blocking call that starts a http server exposing the api
func (api *API) Serve() error {
	server := http.Server{
		Addr:    api.addr,
		Handler: api.router,
	}
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.ListenAndServe()
	}()
	for {
		select {
		case err := <-errCh:
			return err
		case <-api.ctx.Done():
			return server.Close()
		}
	}
}
