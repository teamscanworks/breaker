#! /bin/bash

NAME="breaker"
CHAIN_ID="1234"
KEY_NAME="breaker_admin"
DATA_DIR="$HOME/.simapp"
if [[ "$1" == "reset" ]]; then
    rm -rf "$DATA_DIR"
    simd comet unsafe-reset-all --home "$DATA_DIR"
    simd init "$NAME" --chain-id "$CHAIN_ID" --home "$DATA_DIR" --overwrite
    simd keys add "$KEY_NAME" --home "$DATA_DIR"
    simd genesis add-genesis-account  "$KEY_NAME" 10000000000000000stake --home "$DATA_DIR"
    simd genesis gentx "$KEY_NAME" 10000000000000000stake --chain-id "$CHAIN_ID" --home "$DATA_DIR" --yes --details "breakerchain" --security-contact "breakerchain"  --website "breakerchain" --moniker "$NAME" --amount="10000000000000000stake"
    simd genesis collect-gentxs --home "$DATA_DIR"
    simd genesis validate --home "$DATA_DIR"
else
    simd start --rpc.laddr tcp://127.0.0.1:26657 --home "$DATA_DIR" --x-crisis-skip-assert-invariants --moniker "$NAME" --rpc.unsafe --api.enable --log_level debug --grpc-web.enable
fi
