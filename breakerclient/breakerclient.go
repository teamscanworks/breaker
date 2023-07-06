package breakerclient

import (
	"context"
	"fmt"

	"cosmossdk.io/x/circuit"
	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/pflag"
	compass "github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

// cosmos client that interacts with the x/circuit module, wrapping the compass client
//
// note: uses the very first keypair return from Client.Keyring.List() as the signing keypair
type BreakerClient struct {
	ctx       context.Context
	cancelFn  context.CancelFunc
	Client    *compass.Client
	qc        types.QueryClient
	flagSet   *pflag.FlagSet
	txFactory tx.Factory
	log       *zap.Logger
	cCtx      client.Context
}

// wraps the compass client with additional functionality specific to the x/circuit module
func NewBreakerClient(
	ctx context.Context,
	log *zap.Logger,
	cfg *compass.ClientConfig,
) (*BreakerClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	// set default modules, and register the circuit breaker type
	cfg.Modules = compass.ModuleBasics
	cfg.Modules = append(cfg.Modules, circuit.AppModuleBasic{})
	cl, err := compass.NewClient(log, cfg, []keyring.Option{compass.DefaultSignatureOptions()})
	if err != nil {
		cancel()
		return nil, err
	}
	qc := types.NewQueryClient(cl.GRPC)
	txFactory := cl.TxFactory()
	cCtx := cl.ClientContext()

	bc := &BreakerClient{
		ctx:       ctx,
		cancelFn:  cancel,
		Client:    cl,
		qc:        qc,
		flagSet:   pflag.NewFlagSet("", pflag.ExitOnError),
		txFactory: txFactory,
		log:       log.Named("breaker.client"),
		cCtx:      cCtx,
	}
	if err = bc.SetFromAddress(); err != nil {
		return nil, err
	}
	return bc, nil
}

// helper function that attempts to set the address used by the client context for signing transactions
// logs a warning if no keys are configured, otherwise takes the first available key
func (bc *BreakerClient) SetFromAddress() error {
	keys, err := bc.Client.Keyring.List()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		granter, err := keys[0].GetAddress()
		if err != nil {
			return err
		}
		bc.cCtx = bc.cCtx.WithFromAddress(granter)
		bc.log.Info("configured from address", zap.String("from.address", granter.String()))
	} else {
		bc.log.Warn("no keys found, you should create at least one")
	}
	return nil
}

// lists commands/urls that have had their circuits tripped
func (bc *BreakerClient) ListDisabledCommands(ctx context.Context) (*types.DisabledListResponse, error) {
	return bc.qc.DisabledList(ctx, &types.QueryDisabledListRequest{})
}

// list permissions granted to the given address
func (bc *BreakerClient) Account(ctx context.Context, address string) (*types.AccountResponse, error) {
	return bc.qc.Account(ctx, &types.QueryAccountRequest{Address: address})
}

// returns a paginated list of all accounts that have permissions granted to them
func (bc *BreakerClient) Accounts(ctx context.Context) (*types.AccountsResponse, error) {
	page, err := client.ReadPageRequest(bc.flagSet)
	if err != nil {
		return nil, err
	}
	return bc.qc.Accounts(ctx, &types.QueryAccountsRequest{Pagination: page})
}

// authorize a given account with the specific permission level
func (bc *BreakerClient) Authorize(ctx context.Context, grantee string, permissionLevel string, limitTypeUrls []string) error {
	val, ok := types.Permissions_Level_value[permissionLevel]
	if !ok {
		return fmt.Errorf("failed to find permission level value for key %s", permissionLevel)
	}
	permission := types.Permissions{
		Level:         types.Permissions_Level(val),
		LimitTypeUrls: limitTypeUrls,
	}
	granterAddr := bc.cCtx.GetFromAddress().String()
	msg := types.NewMsgAuthorizeCircuitBreaker(granterAddr, grantee, &permission)
	if err := tx.BroadcastTx(bc.cCtx, bc.txFactory, msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// trip a circuit for the given urls, preventing calls to those module requests
func (bc *BreakerClient) TripCircuitBreaker(ctx context.Context, urls []string) error {
	granterAddr := bc.cCtx.GetFromAddress().String()
	msg := types.NewMsgTripCircuitBreaker(granterAddr, urls)
	if err := tx.BroadcastTx(bc.cCtx, bc.txFactory, msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// resets a tripped circuit, allowing calls to those module requests again
func (bc *BreakerClient) ResetCircuitBreaker(ctx context.Context, urls []string) error {
	granterAddr := bc.cCtx.GetFromAddress().String()
	msg := types.NewMsgResetCircuitBreaker(granterAddr, urls)
	if err := tx.BroadcastTx(bc.cCtx, bc.txFactory, msg); err != nil {
		bc.log.Error("transaction broadcast failed", zap.Error(err), zap.Stack("stacktrace"))
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

// creates a new mnemonic phrase and inserts into the keyring
func (bc *BreakerClient) NewMnemonic(keyName string, mnemonic ...string) (string, error) {
	keyOutput, err := bc.Client.KeyAddOrRestore(keyName, 118, mnemonic...)
	if err != nil {
		bc.log.Error("failed to add new key", zap.Error(err))
		return "", fmt.Errorf("failed to create new mnemonic %s", err)
	}
	return keyOutput.Mnemonic, nil
}
