#! /bin/bash

source "$PWD/scripts/log.sh"

info_log "voting yes from breaker_admin"
simd tx gov vote 1 yes --from breaker_admin
sleep 2
info_log "voting yes from breaker_voter"
simd tx gov vote 1 yes --from breaker_voter
info_log "sleeping for 70 seconds to allow proposal finalization"
