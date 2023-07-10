.PHONY: build
build:
	go build -o breaker-cli .
	
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


# Static analysis and style checks
.PHONY: lint
lint:
	go fmt ./...
	go vet ./...
	
# Execute all tests
.PHONY: test
test:
	@echo "===================   executing all tests   ==================="
	go test -cover ./...
	@echo "===================          done           ==================="