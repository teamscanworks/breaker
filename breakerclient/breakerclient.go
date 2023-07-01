package breakerclient

import (
	"context"
	"fmt"

	"github.com/spf13/pflag"
	lens "github.com/teamscanworks/breaker/client"

	"cosmossdk.io/x/circuit/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	config "github.com/teamscanworks/breaker/config"
)

type BreakerClient struct {
	ctx       context.Context
	cancel    context.CancelFunc
	clientCtx client.Context
	lc        *lens.ChainClient
	qc        types.QueryClient
	fc        tx.Factory
	flagSet   *pflag.FlagSet
}

func NewBreakerClient(
	ctx context.Context,
	cfg config.Config,
) (*BreakerClient, error) {
	ctx, cancel := context.WithCancel(ctx)
	logger, err := cfg.ZapLogger()
	if err != nil {
		return nil, err
	}
	chainConfig := cfg.ChainClientConfig()
	lc, err := cfg.ChainClient(logger, &chainConfig)
	if err != nil {
		return nil, err
	}
	fc := lc.TxFactory()
	//rc, err := client.NewClientFromNode(cfg.Cosmos.RPCEndpoint)
	//if err != nil {
	//	return nil, err
	//}
	//grpcConn, err := grpc.Dial(
	//	cfg.Cosmos.GRPCEndpoint, // your gRPC server address.
	//	grpc.WithInsecure(),     // The Cosmos SDK doesn't support any transport security mechanism.
	//	// This instantiates a general gRPC codec which handles proto bytes. We pass in a nil interface registry
	//	// if the request/response types contain interface instead of 'nil' you should pass the application specific codec.
	//	grpc.WithDefaultCallOptions(grpc.ForceCodec(codec.NewProtoCodec(nil).GRPCCodec())),
	//)
	//if err != nil {
	//	return nil, err
	//}
	flagSet := pflag.NewFlagSet("", pflag.ExitOnError)
	clientCtx := cfg.ClientContext()
	clientCtx = clientCtx.WithClient(lc.RPCClient)
	//ctx, err = client.ReadPersistentCommandFlags(ctx, flagSet)
	//if err != nil {
	//	return nil, err
	//}
	//ctx = ctx.WithClient(rc)
	//ctx = ctx.WithGRPCClient(grpcConn)
	qc := types.NewQueryClient(clientCtx)

	return &BreakerClient{
		ctx,
		cancel,
		clientCtx,
		lc,
		qc,
		fc,
		flagSet,
	}, nil
}

func (bc *BreakerClient) Close() error {
	bc.cancel()
	return nil
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
	msg := types.NewMsgAuthorizeCircuitBreaker(bc.clientCtx.GetFromAddress().String(), grantee, &permission)
	if err := tx.BroadcastTx(bc.clientCtx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) TripCircuitBreaker(ctx context.Context, urls []string) error {
	msg := types.NewMsgTripCircuitBreaker(bc.clientCtx.GetFromAddress().String(), urls)
	if err := tx.BroadcastTx(bc.clientCtx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}

func (bc *BreakerClient) ResetCircuitBreaker(ctx context.Context, urls []string) error {
	msg := types.NewMsgResetCircuitBreaker(bc.clientCtx.GetFromAddress().String(), urls)
	if err := tx.BroadcastTx(bc.clientCtx, bc.fc, msg); err != nil {
		return fmt.Errorf("failed to broadcast transaction %v", err)
	}
	return nil
}
