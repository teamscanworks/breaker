#! /bin/bash

source "$PWD/scripts/log.sh"

NAME="breaker"
CHAIN_ID="testing"
KEY_NAME="breaker_admin"
VOTER_NAME="breaker_voter"

TEST_ACCOUNT="cosmos18q2gyed58368mmrkz3k30s6kyrx0p4wrykals7"
if [[ "$1" == "reset" ]]; then
    info_log "removing old state"
    simd comet unsafe-reset-all
    rm -rf "$HOME/.simapp"
    
    info_log "initializing simd"
    simd init "$NAME" --chain-id "$CHAIN_ID"
    
    # preload the keys used for testing
    echo "resource ridge evolve huge forum train category curtain elegant valley disorder idea elder tenant belt sibling spin little athlete range media syrup rigid poem" | simd keys add "$KEY_NAME" --recover
    echo "off notice dress fantasy type cargo among jaguar cream ride swift shuffle wear below citizen trim worry huge fire champion tunnel unique please wine" | simd keys add "$VOTER_NAME" --recover

    info_log "adding genesis account for $TEST_ACCOUNT"
    simd genesis add-genesis-account "$TEST_ACCOUNT" 1stake
    
    info_log "adding genesis account for $VOTER_NAME"
    simd genesis add-genesis-account "$VOTER_NAME" 10000000000000000stake
    
    info_log "adding genesis account for $KEY_NAME"
    simd genesis add-genesis-account "$KEY_NAME" 20000000000000000stake
    
    info_log "registering validator for $KEY_NAME"
    simd genesis gentx "$KEY_NAME" 19000000000000000stake --chain-id "$CHAIN_ID" --yes --details "breakerchain" --security-contact "breakerchain"  --website "breakerchain" --amount="19000000000000000stake" #--keyring-backend test 

    info_log "collecting genesis txs"
    simd genesis collect-gentxs

    info_log "overriding default genesis parameters"
    sed -i 's/86400s/60s/g' $HOME/.simapp/config/genesis.json
    sed -i 's/timeout_commit = \"5s\"/timeout_commit = \"2s\"/g' $HOME/.simapp/config/config.toml
    sed -i 's/0.667000000000000000/0.500001/g' $HOME/.simapp/config/genesis.json

    # to change the voting_period
    jq '.app_state.gov.params.voting_period = "65s"' $HOME/.simapp/config/genesis.json > temp.json && mv temp.json $HOME/.simapp/config/genesis.json
    jq '.app_state.gov.params.quorum = "0.01"' $HOME/.simapp/config/genesis.json > temp.json && mv temp.json $HOME/.simapp/config/genesis.json
    jq '.app_state.gov.params.threshold = "0.05"' $HOME/.simapp/config/genesis.json > temp.json && mv temp.json $HOME/.simapp/config/genesis.json

    # to change the inflation
    jq '.app_state.mint.minter.inflation = "0.300000000000000000"' $HOME/.simapp/config/genesis.json > temp.json && mv temp.json $HOME/.simapp/config/genesis.json

    info_log "validating genesis file"
    simd genesis validate --log_level warn
else
    simd start --rpc.laddr tcp://127.0.0.1:26657 --x-crisis-skip-assert-invariants --rpc.unsafe --api.enable --grpc-web.enable --api.swagger --trace --trace-store trace_data.txt --minimum-gas-prices 0stake
fi
