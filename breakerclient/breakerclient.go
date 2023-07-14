package breakerclient

import (
	"context"
	"fmt"

	"cosmossdk.io/x/circuit"
	"cosmossdk.io/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/pflag"
	compass "github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

// cosmos client that interacts with the x/circuit module, wrapping the compass client
// in order to send transactions with BreakerClient, you should configure `BreakerClient::Client::Keyring`
// with at least one key after `NewBreakerClient` returns, followed by `BreakerClient::SetFromAddress`
// ```
// bc, err := breakerclient.NewBreakerClient(ctx, log, cfg)
//
//	if err != nil {
//		panic(err)
//	}
//
// bc.NewMnemonic("defaultkey")
// err = bc.SetFromAddress()
//
//	if err != nil {
//		panic(err)
//	}
//
// ````
type BreakerClient struct {
	Client   *compass.Client
	flagSet  *pflag.FlagSet
	log      *zap.Logger
	qc       types.QueryClient
	ctx      context.Context
	cancelFn context.CancelFunc
}

// Wraps the compass client with additional functionality specific to the x/circuit module.
func NewBreakerClient(
	ctx context.Context,
	log *zap.Logger,
	cfg *compass.ClientConfig,
) (*BreakerClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	// set default modules, and register the circuit breaker type
	cfg.Modules = compass.ModuleBasics
	cfg.Modules = append(cfg.Modules, circuit.AppModuleBasic{})
	// initialize compass client with default signature list
	cl, err := compass.NewClient(log, cfg, []keyring.Option{compass.DefaultSignatureOptions()})
	if err != nil {
		cancel()
		return nil, err
	}
	// initialize circuit breaker module specific clients
	qc := types.NewQueryClient(cl.GRPC)

	bc := &BreakerClient{
		ctx:      ctx,
		cancelFn: cancel,
		Client:   cl,
		qc:       qc,
		flagSet:  pflag.NewFlagSet("", pflag.ExitOnError),
		log:      log.Named("breaker.client"),
	}
	// attempt to set a from account if there is at least 1 key in the keyring
	if err = bc.SetFromAddress(); err != nil {
		cancel()
		return nil, err
	}
	return bc, nil
}

// Helper function that attempts to set the address used by the client context for signing transactions
// logs a warning if no keys are configured, otherwise takes the first available key.
func (bc *BreakerClient) SetFromAddress() error {
	return bc.Client.SetFromAddress()
}

// Returns the keypair actively in use for signing transactions (the first key in the keyring).
// If no address has been configured returns `nil, nil`
func (bc *BreakerClient) GetActiveKeypair() (*sdktypes.AccAddress, error) {
	return bc.Client.GetActiveKeypair()
}

// Lists commands/urls that have had their circuits tripped.
func (bc *BreakerClient) ListDisabledCommands(ctx context.Context) (*types.DisabledListResponse, error) {
	return bc.qc.DisabledList(ctx, &types.QueryDisabledListRequest{})
}

// List permissions granted to the given address.
func (bc *BreakerClient) Account(ctx context.Context, address string) (*types.AccountResponse, error) {
	return bc.qc.Account(ctx, &types.QueryAccountRequest{Address: address})
}

// Returns a paginated list of all accounts that have permissions granted to them.
func (bc *BreakerClient) Accounts(ctx context.Context) (*types.AccountsResponse, error) {
	page, err := client.ReadPageRequest(bc.flagSet)
	if err != nil {
		return nil, err
	}
	return bc.qc.Accounts(ctx, &types.QueryAccountsRequest{Pagination: page})
}

// Authorize a given account with the specific permission level.
func (bc *BreakerClient) Authorize(ctx context.Context, grantee string, permissionLevel string, limitTypeUrls []string) error {
	val, ok := types.Permissions_Level_value[permissionLevel]
	if !ok {
		return fmt.Errorf("failed to find permission level value for key %s", permissionLevel)
	}
	permission := types.Permissions{
		Level:         types.Permissions_Level(val),
		LimitTypeUrls: limitTypeUrls,
	}
	granterAddr := bc.ClientContext().GetFromAddress().String()
	msg := types.NewMsgAuthorizeCircuitBreaker(granterAddr, grantee, &permission)
	if err := tx.BroadcastTx(bc.ClientContext(), bc.TxFactory(), msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// Trip a circuit for the given urls, preventing calls to the module request urls.
func (bc *BreakerClient) TripCircuitBreaker(ctx context.Context, urls []string) error {
	granterAddr := bc.ClientContext().GetFromAddress().String()
	msg := types.NewMsgTripCircuitBreaker(granterAddr, urls)
	bc.log.Debug("tripping breaker", zap.Any("message", msg), zap.String("granter.addr", granterAddr))
	if err := tx.BroadcastTx(bc.ClientContext(), bc.TxFactory(), msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// Resets a tripped circuit, allowing calls to the module request urls.
func (bc *BreakerClient) ResetCircuitBreaker(ctx context.Context, urls []string) error {
	granterAddr := bc.ClientContext().GetFromAddress().String()
	msg := types.NewMsgResetCircuitBreaker(granterAddr, urls)
	if err := tx.BroadcastTx(bc.ClientContext(), bc.TxFactory(), msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// Creates a new mnemonic phrase and inserts into the configured keyring. Coin type defaults to 118.
func (bc *BreakerClient) NewMnemonic(keyName string, mnemonic ...string) (string, error) {
	keyOutput, err := bc.Client.KeyAddOrRestore(keyName, 118, mnemonic...)
	if err != nil {
		bc.log.Error("failed to add new key", zap.Error(err))
		return "", fmt.Errorf("failed to create new mnemonic %s", err)
	}
	return keyOutput.Mnemonic, nil
}

func (bc *BreakerClient) Prepare() error {
	return bc.Client.PrepareClientContext(bc.ClientContext())
}

func (bc *BreakerClient) ClientContext() client.Context {
	return bc.Client.ClientContext()
}

func (bc *BreakerClient) TxFactory() tx.Factory {
	return bc.Client.TxFactory()
}
