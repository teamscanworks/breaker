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