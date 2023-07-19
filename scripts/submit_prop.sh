#! /bin/bash

source "$PWD/scripts/log.sh"

BREAKER_ADDR=$(simd query staking validators | grep operator_address | awk '{print $NF}')
VOTER_ADDR=$(simd keys show breaker_voter --address)
info_log "delegating stake from breaker_voter to breaker_admin"
simd tx staking delegate "$BREAKER_ADDR"  9000000000000000stake --from "$VOTER_ADDR"
info_log "submitting draft proposal from breaker_admin"
simd tx gov submit-proposal scripts/draft_proposal.json --from breaker_admin
sleep 2