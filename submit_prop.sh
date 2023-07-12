#! /bin/bash

simd tx gov submit-proposal draft_proposal.json --from breaker_admin
sleep 2
simd tx gov deposit 1 10000000000000000stake --from breaker_voter
sleep 2
simd tx gov vote 1 yes --from breaker_admin
sleep 2
simd tx gov deposit 1 10000000000000000stake --from breaker_voter
sleep 2
simd tx gov vote 1 yes --from breaker_voter
