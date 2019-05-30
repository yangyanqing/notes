module main.go

go 1.12

require (
	github.com/cosmos/cosmos-sdk v0.34.5
	github.com/syndtr/goleveldb v1.0.0
	golang.org/x/crypto v0.0.0-20181203042331-505ab145d0a9 // indirect
)

replace golang.org/x/crypto => github.com/tendermint/crypto v0.0.0-20180820045704-3764759f34a5
