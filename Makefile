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

.PHONY: example-authorize
example-authorizer:
	simd tx circuit authorize cosmos1lvdt3jkkde8ppn94m4w54zxr065rxzsc35tjnm 0 "cosmos.bank.v1beta1.MsgSend,cosmos.bank.v1beta1.MsgMultiSend" --from cosmos1lvdt3jkkde8ppn94m4w54zxr065rxzsc35tjnm --chain-id 1234

# Static analysis and style checks
.PHONY: lint
lint:
	go fmt ./...
	go vet ./...

# Execute short tests
.PHONY: test
test:
	@echo "===================  executing short tests  ==================="
	go test -race -cover -short ./...
	@echo "===================          done           ==================="

# Execute all tests
.PHONY: test
test-all:
	@echo "===================   executing all tests   ==================="
	go test -race -cover ./...
	@echo "===================          done           ==================="