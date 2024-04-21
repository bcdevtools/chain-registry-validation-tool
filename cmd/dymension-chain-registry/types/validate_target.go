package types

import "strings"

type ValidateTarget string

const (
	ValidateMainnet        ValidateTarget = "mainnet"
	ValidateTestnet        ValidateTarget = "testnet"
	ValidateDevnet         ValidateTarget = "devnet"
	ValidateInternalDevnet ValidateTarget = "internal-devnet"
)

func (vt ValidateTarget) String() string {
	switch vt {
	case ValidateInternalDevnet:
		return "Internal Devnet"
	default:
		return strings.ToUpper(string(string(vt)[0])) + string(vt)[1:]
	}
}

func (vt ValidateTarget) SubDirectoryName() string {
	return string(vt)
}
