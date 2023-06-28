package breakerclient

import (
	"context"
	"fmt"

	"cosmossdk.io/x/circuit/types"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/pflag"
	config "github.com/teamscanworks/breaker/config"
)

type BreakerClient struct {
	ctx client.Context
	rc  *rpchttp.HTTP
	qc  types.QueryClient
	fc  tx.Factory
}

func NewBreakerClient(
	cfg config.Config,
) (*BreakerClient, error) {
	rc, err := client.NewClientFromNode(cfg.Cosmos.RPCEndpoint)
	if err != nil {
		return nil, err
	}

	flagSet := pflag.NewFlagSet("", pflag.ExitOnError)
	ctx := cfg.ClientContext()
	ctx, err = client.ReadPersistentCommandFlags(ctx, flagSet)
	if err != nil {
		return nil, err
	}
	ctx = ctx.WithClient(rc)
	qc := types.NewQueryClient(ctx)
	fc, err := tx.NewFactoryCLI(ctx, flagSet)
	if err != nil {
		return nil, err
	}
	return &BreakerClient{
		ctx,
		rc,
		qc,
		fc,
	}, nil
}

func (bc *BreakerClient) ListDisabledCommands(ctx context.Context) (*types.DisabledListResponse, error) {
	return bc.qc.DisabledList(ctx, &types.QueryDisabledListRequest{})
}

func (bc *BreakerClient) Account(ctx context.Context, address string) (*types.AccountResponse, error) {
	return bc.qc.Account(ctx, &types.QueryAccountRequest{Address: address})
}

func (bc *BreakerClient) Accounts(ctx context.Context) (*types.AccountsResponse, error) {
	page, err := client.ReadPageRequest(nil)
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
	msg := types.NewMsgAuthorizeCircuitBreaker(bc.ctx.GetFromAddress().String(), grantee, &permission)
	if err := tx.BroadcastTx(bc.ctx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) TripCircuitBreaker(ctx context.Context, urls []string) error {
	msg := types.NewMsgTripCircuitBreaker(bc.ctx.GetFromAddress().String(), urls)
	if err := tx.BroadcastTx(bc.ctx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) ResetCircuitBreaker(ctx context.Context, urls []string) error {
	msg := types.NewMsgResetCircuitBreaker(bc.ctx.GetFromAddress().String(), urls)
	if err := tx.BroadcastTx(bc.ctx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}
