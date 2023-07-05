package breakerclient

import (
	"context"
	"fmt"

	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/pflag"
	compass "github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

type BreakerClient struct {
	ctx       context.Context
	cancelFn  context.CancelFunc
	Client    *compass.Client
	qc        types.QueryClient
	flagSet   *pflag.FlagSet
	txFactory tx.Factory
}

func NewBreakerClient(
	ctx context.Context,
	log *zap.Logger,
	cfg *compass.ClientConfig,
) (*BreakerClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	cl, err := compass.NewClient(log, cfg)
	if err != nil {
		cancel()
		return nil, err
	}
	qc := types.NewQueryClient(cl.GRPC)
	txFactory := cl.TxFactory()
	return &BreakerClient{
		ctx:       ctx,
		cancelFn:  cancel,
		Client:    cl,
		qc:        qc,
		flagSet:   pflag.NewFlagSet("", pflag.ExitOnError),
		txFactory: txFactory,
	}, nil
}

func (bc *BreakerClient) ListDisabledCommands(ctx context.Context) (*types.DisabledListResponse, error) {
	return bc.qc.DisabledList(ctx, &types.QueryDisabledListRequest{})
}

func (bc *BreakerClient) Account(ctx context.Context, address string) (*types.AccountResponse, error) {
	return bc.qc.Account(ctx, &types.QueryAccountRequest{Address: address})
}

func (bc *BreakerClient) Accounts(ctx context.Context) (*types.AccountsResponse, error) {
	page, err := client.ReadPageRequest(bc.flagSet)
	if err != nil {
		return nil, err
	}
	return bc.qc.Accounts(ctx, &types.QueryAccountsRequest{Pagination: page})
}

func (bc *BreakerClient) Authorize(ctx context.Context, grantee string, permissionLevel string, limitTypeUrls []string) error {
	val, ok := types.Permissions_Level_value[permissionLevel]
	if !ok {
		return fmt.Errorf("failed to find permission level value for key %s", permissionLevel)
	}
	permission := types.Permissions{
		Level:         types.Permissions_Level(val),
		LimitTypeUrls: limitTypeUrls,
	}
	keys, err := bc.Client.Keyring.List()
	if err != nil {
		return fmt.Errorf("failed to list keyring %s", err)
	}
	granter := keys[0]
	granterAddr, err := granter.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to get address %s", err)
	}
	msg := types.NewMsgAuthorizeCircuitBreaker(granterAddr.String(), grantee, &permission)
	if err := tx.BroadcastTx(bc.Client.ClientContext(), bc.txFactory, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) TripCircuitBreaker(ctx context.Context, urls []string) error {
	keys, err := bc.Client.Keyring.List()
	if err != nil {
		return fmt.Errorf("failed to list keyring %s", err)
	}
	granter := keys[0]
	granterAddr, err := granter.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to get address %s", err)
	}
	msg := types.NewMsgTripCircuitBreaker(granterAddr.String(), urls)
	if err := tx.BroadcastTx(bc.Client.ClientContext(), bc.txFactory, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) ResetCircuitBreaker(ctx context.Context, urls []string) error {
	keys, err := bc.Client.Keyring.List()
	if err != nil {
		return fmt.Errorf("failed to list keyring %s", err)
	}
	granter := keys[0]
	granterAddr, err := granter.GetAddress()
	if err != nil {
		return fmt.Errorf("failed to get address %s", err)
	}
	msg := types.NewMsgResetCircuitBreaker(granterAddr.String(), urls)
	if err := tx.BroadcastTx(bc.Client.ClientContext(), bc.txFactory, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}
