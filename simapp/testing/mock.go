package testing

import (
	"encoding/json"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmcrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CosmWasm/wasmd/x/wasm"

	"github.com/larry0x/abstract-account/simapp"
	poatypes "github.com/larry0x/simapp/x/poa/types"
)

const DefaultBondDenom = "utoken"

var DefaultConsensusParams = &tmproto.ConsensusParams{
	Block: &tmproto.BlockParams{
		MaxBytes: 200000,
		MaxGas:   10000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

func MakeSimpleMockApp() *simapp.SimApp {
	return MakeMockApp([]banktypes.Balance{})
}

func MakeMockApp(balances []banktypes.Balance) *simapp.SimApp {
	encCfg := simapp.MakeEncodingConfig()

	app := simapp.NewSimApp(
		log.NewNopLogger(),
		dbm.NewMemDB(),
		nil,
		true,
		EmptyAppOptions{},
		[]wasm.Option{},
	)

	gs := MakeMockGenesisState(encCfg.Codec, balances)
	gsBytes, err := json.Marshal(gs)
	if err != nil {
		panic(err)
	}

	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   gsBytes,
		},
	)

	return app
}

func MakeMockGenesisState(cdc codec.JSONCodec, balances []banktypes.Balance) simapp.GenesisState {
	gs := simapp.DefaultGenesisState(cdc)

	// prepare genesis accounts
	genAccts := []authtypes.GenesisAccount{}
	for _, balance := range balances {
		addr, err := sdk.AccAddressFromBech32(balance.Address)
		if err != nil {
			panic(err)
		}

		genAccts = append(genAccts, authtypes.NewBaseAccountWithAddress(addr))
	}

	// compute total supply of tokens
	supply := sdk.NewCoins()
	for _, balance := range balances {
		supply = supply.Add(balance.Coins...)
	}

	gs[authtypes.ModuleName] = cdc.MustMarshalJSON(authtypes.NewGenesisState(
		authtypes.DefaultParams(),
		genAccts,
	))

	gs[banktypes.ModuleName] = cdc.MustMarshalJSON(banktypes.NewGenesisState(
		banktypes.DefaultParams(),
		balances,
		supply,
		[]banktypes.Metadata{},
		[]banktypes.SendEnabled{},
	))

	gs[poatypes.ModuleName] = cdc.MustMarshalJSON(&poatypes.GenesisState{
		Validators: []abci.ValidatorUpdate{
			{
				PubKey: tmcrypto.PublicKey{
					Sum: &tmcrypto.PublicKey_Ed25519{
						Ed25519: MakeRandomConsensusPubKey().Bytes(),
					},
				},
				Power: 1,
			},
		},
	})

	return gs
}

// ----------------------------------- Keys ------------------------------------

func MakeRandomAddress() sdk.AccAddress {
	return sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func MakeRandomPubKey() cryptotypes.PubKey {
	return secp256k1.GenPrivKey().PubKey()
}

func MakeRandomConsensusPubKey() cryptotypes.PubKey {
	return ed25519.GenPrivKey().PubKey()
}

// -------------------------------- AppOptions ---------------------------------

type EmptyAppOptions struct{}

func (opts EmptyAppOptions) Get(_ string) interface{} {
	return nil
}
