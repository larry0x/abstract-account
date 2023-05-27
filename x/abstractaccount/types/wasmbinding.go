package types

import "github.com/cosmos/gogoproto/proto"

type Any struct {
	TypeURL string `json:"type_url"`
	Value   []byte `json:"value"`
}

type AccountSudoMsg struct {
	BeforeTx *BeforeTx `json:"before_tx,omitempty"`
	AfterTx  *AfterTx  `json:"after_tx,omitempty"`
}

type BeforeTx struct {
	Msgs       []*Any `json:"msgs"`
	TxBytes    []byte `json:"tx_bytes"`
	Credential []byte `json:"credential"`
}

type AfterTx struct {
	Success bool `json:"success"`
}

func NewAnyFromProtoMsg(msg proto.Message) (*Any, error) {
	bz, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	any := &Any{
		TypeURL: "/" + proto.MessageName(msg),
		Value:   bz,
	}

	return any, nil
}
