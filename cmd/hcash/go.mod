module github.com/hpb-project/HCash-SDK/cmd/hcash

go 1.15

require (
	github.com/ethereum/go-ethereum v1.10.3
	github.com/hpb-project/HCash-SDK v0.0.6
)
replace (
	github.com/hpb-project/HCash-SDK => ../../
)