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
	Msgs      []*Any `json:"msgs"`
	TxBytes   []byte `json:"tx_bytes"`
	CredBytes []byte `json:"cred_bytes,omitempty"`
	Simulate  bool   `json:"simulate"`
}

type AfterTx struct {
	Simulate bool `json:"simulate"`
}

func NewAnyFromProtoMsg(msg proto.Message) (*Any, error) {
	msgBytes, err := proto.Marshal(msg)
	if err != nil {
		return nil, err
	}

	msgAny := &Any{
		TypeURL: "/" + proto.MessageName(msg),
		Value:   msgBytes,
	}

	return msgAny, nil
}
