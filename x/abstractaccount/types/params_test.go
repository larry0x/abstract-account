package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func TestValidateParams(t *testing.T) {
	for _, tc := range []struct {
		params *types.Params
		expErr bool
	}{
		{
			params: &types.Params{},
			expErr: true,
		},
		{
			params: &types.Params{MaxGasBefore: 0, MaxGasAfter: types.DefaultMaxGas},
			expErr: true,
		},
		{
			params: &types.Params{MaxGasBefore: types.DefaultMaxGas, MaxGasAfter: 0},
			expErr: true,
		},
		{
			params: &types.Params{MaxGasBefore: types.DefaultMaxGas, MaxGasAfter: types.DefaultMaxGas},
			expErr: false,
		},
	} {
		err := tc.params.Validate()

		if tc.expErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
