#! /bin/bash

NAME="breaker"
CHAIN_ID="testing"
KEY_NAME="breaker_admin"
VOTER_NAME="breaker_voter"
#VALIDATOR_KEY="breaker_validator"
TEST_ACCOUNT="cosmos18q2gyed58368mmrkz3k30s6kyrx0p4wrykals7"
if [[ "$1" == "reset" ]]; then
    echo "[INFO] removing old state"
    simd comet unsafe-reset-all
    rm -rf "$HOME/.simapp"
    echo "[INFO] initializing simd"
    simd init "$NAME" --chain-id "$CHAIN_ID"
    echo "resource ridge evolve huge forum train category curtain elegant valley disorder idea elder tenant belt sibling spin little athlete range media syrup rigid poem" | simd keys add "$KEY_NAME" --recover
    echo "off notice dress fantasy type cargo among jaguar cream ride swift shuffle wear below citizen trim worry huge fire champion tunnel unique please wine" | simd keys add "$VOTER_NAME" --recover
    echo "[INFO] adding genesis account for $TEST_ACCOUNT"
    simd genesis add-genesis-account "$TEST_ACCOUNT" 1stake
    echo "[INFO] adding genesis account for $VOTER_NAME"
    simd genesis add-genesis-account "$VOTER_NAME" 20000000000000000stake
    echo "[INFO] adding genesis account for $KEY_NAME"
    simd genesis add-genesis-account "$KEY_NAME" 20000000000000000stake #--keyring-backend test 
    #simd genesis add-genesis-account $VALIDATOR_KEY 10000000000000000stake #--keyring-backend test 
#    simd genesis add-genesis-account gov 1atom --module-name gov
    #simd genesis collect-gentxs
    simd genesis gentx "$KEY_NAME" 100000000000stake --chain-id "$CHAIN_ID" --yes --details "breakerchain" --security-contact "breakerchain"  --website "breakerchain" --amount="100000000000stake" #--keyring-backend test 
    #simd genesis collect-gentxs --home "$DATA_DIR"
    simd genesis collect-gentxs
    # overwrite the expedited voting time to 60s
    sed -i 's/86400s/60s/g' $HOME/.simapp/config/genesis.json
    sed -i 's/timeout_commit = \"5s\"/timeout_commit = \"2s\"/g' $HOME/.simapp/config/config.toml
    sed -i 's/0.667000000000000000/0.500001/g' $HOME/.simapp/config/genesis.json
    simd genesis validate
else
    simd start --rpc.laddr tcp://127.0.0.1:26657 --x-crisis-skip-assert-invariants --rpc.unsafe --api.enable --grpc-web.enable --api.swagger --trace --trace-store trace_data.txt --minimum-gas-prices 0stake
fi
