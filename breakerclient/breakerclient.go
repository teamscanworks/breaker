package breakerclient

import (
	"context"

	"cosmossdk.io/x/circuit/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/pflag"
	compass "github.com/teamscanworks/compass"
	"go.uber.org/zap"
)

type BreakerClient struct {
	ctx      context.Context
	cancelFn context.CancelFunc
	Client   *compass.Client
	qc       types.QueryClient
	flagSet  *pflag.FlagSet
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
	return &BreakerClient{
		ctx:      ctx,
		cancelFn: cancel,
		Client:   cl,
		qc:       qc,
		flagSet:  pflag.NewFlagSet("", pflag.ExitOnError),
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
