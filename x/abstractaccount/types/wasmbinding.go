package types

import wasmvmtypes "github.com/CosmWasm/wasmvm/types"

type AccountSudoMsg struct {
	BeforeTx *BeforeTx `json:"before_tx,omitempty"`
	AfterTx  *AfterTx  `json:"after_tx,omitempty"`
}

type BeforeTx struct {
	Msgs      []wasmvmtypes.StargateMsg `json:"msgs"`
	SignBytes []byte                    `json:"sign_bytes"`
	Signature []byte                    `json:"signature"`
}

type AfterTx struct {
	Success bool `json:"success"`
}
