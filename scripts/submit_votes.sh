#! /bin/bash

echo "[INFO] voting yes from breaker_admin"
simd tx gov vote 1 yes --from breaker_admin
sleep 2
echo "[INFO] voting yes from breaker_voter"
simd tx gov vote 1 yes --from breaker_voter
