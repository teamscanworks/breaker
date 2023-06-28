.PHONY: install-simd
install-simd:
	git clone https://github.com/cosmos/cosmos-sdk
	cd cosmos-sdk
	make build

.PHONY: start-simd
start-simd:
	./scripts/start_simd.sh

.PHONY: reset-simd
reset-simd:
	./scripts/start_simd.sh reset