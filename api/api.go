package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/jwtauth/v5"
	"github.com/teamscanworks/breaker/breakerclient"
	"go.uber.org/zap"
)

// Http api that exposes x/circuit module functionality, primarily used to trip and reset circuits
type API struct {
	ctx           context.Context
	cancel        context.CancelFunc
	router        chi.Router
	logger        *zap.Logger
	jwt           *JWT
	breakerClient *breakerclient.BreakerClient
	addr          string
	// used to block closure until api is shutdown
	doneCh chan struct{}
}

// Options used to configured the API Server
type ApiOpts struct {
	ListenAddress                string
	Password                     string
	IdentifierField              string
	TokenValidityDurationSeconds int64
}

// Prepares the http api server
func NewAPI(
	ctx context.Context,
	log *zap.Logger,
	jwt *JWT,
	opts ApiOpts,
	bc *breakerclient.BreakerClient,
) (*API, error) {
	ctx, cancel := context.WithCancel(ctx)

	api := API{
		ctx:    ctx,
		cancel: cancel,
		router: chi.NewRouter(),
		jwt: NewJWT(
			opts.Password,
			opts.IdentifierField,
			opts.TokenValidityDurationSeconds,
		),
		addr:          opts.ListenAddress,
		logger:        log.Named("breaker.api"),
		breakerClient: bc,
		doneCh:        make(chan struct{}, 1),
	}

	// initialize router
	api.router.Use(middleware.RequestID)
	api.router.Use(NewLoggerMiddleware(api.logger))
	api.router.Route("/v1", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			// authenticated urls
			r.Use(jwtauth.Verifier(api.jwt.tokenAuth))
			r.Use(api.jwt.Authenticator)
			r.Post("/webhook", api.HandleWebookV1)
		})
		r.Group(func(r chi.Router) {
			// unauthenticated urls
			r.Route("/status", func(r chi.Router) {
				r.Route("/list", func(r chi.Router) {
					r.Get("/disabledCommands", api.ListDisabledCommands)
					r.Get("/accounts", api.ListAccounts)
				})
			})
		})
	})

	return &api, nil
}

// Configures the breakerclient such that it may be used by the API for signing transactions.
// This should be called against breakerclient.BreakerClient before passing it as a parameter during api initialization
func ConfigBreakerClient(
	client *breakerclient.BreakerClient,
	keyName string,
) error {
	if err := client.SetFromAddress(); err != nil {
		return fmt.Errorf("failed to initialize from address %s", err)
	}
	client.UpdateClientFromName(keyName)
	return nil
}

// Cancels the api context, triggering a shutdown of the api router.
func (api *API) Close() {
	api.cancel()
	<-api.doneCh
}

// Blocking call that starts a http server exposing the api.
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
			err := server.Close()
			api.doneCh <- struct{}{}
			return err
		}
	}
}
