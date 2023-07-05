#! /bin/bash

NAME="breaker"
CHAIN_ID="1234"
KEY_NAME="breaker_admin"
DATA_DIR="$HOME/.simapp"
ALICE_ADDR="cosmos1cr2u792nntd24mlczu8x4eyfx0tuaa3yqardcv"
VALIDATOR_ADDR="cosmos1xy3zta7m6f90lqg2nfvc47f78s9js2q0dr4q65"
if [[ "$1" == "reset" ]]; then
    rm -rf "$DATA_DIR"
    simd comet unsafe-reset-all
    simd init "$NAME" --chain-id "$CHAIN_ID" --home "$DATA_DIR" --overwrite
    #simd keys add "$KEY_NAME" --home "$DATA_DIR"
    simd genesis add-genesis-account  "$KEY_NAME" 10000000000000000stake #--keyring-backend test 
    simd genesis add-genesis-account $ALICE_ADDR 100000000000stake #--keyring-backend test 
    simd genesis add-genesis-account $VALIDATOR_ADDR 100000000000stake #--keyring-backend test 
    #simd genesis collect-gentxs
    simd genesis gentx "$KEY_NAME" 10000000000000000stake --chain-id "$CHAIN_ID" --yes --details "breakerchain" --security-contact "breakerchain"  --website "breakerchain" --amount="10000000000000000stake" #--keyring-backend test 
    #simd genesis collect-gentxs --home "$DATA_DIR"
    simd genesis collect-gentxs
    simd genesis validate
else
    simd start --rpc.laddr tcp://127.0.0.1:26657 --x-crisis-skip-assert-invariants --rpc.unsafe --api.enable --log_level debug --grpc-web.enable
fi
