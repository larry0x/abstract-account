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

func TestDeterminedAllowedCodeID(t *testing.T) {
	for _, tc := range []struct {
		allowAllCodeIDs bool
		allowedCodeIDs  []uint64
		codeID          uint64
		expAllowed      bool
	}{
		{
			allowAllCodeIDs: true,
			allowedCodeIDs:  []uint64{},
			codeID:          69420,
			expAllowed:      true,
		},
		{
			allowAllCodeIDs: false,
			allowedCodeIDs:  []uint64{12345, 42069, 69420},
			codeID:          69420,
			expAllowed:      true,
		},
		{
			allowAllCodeIDs: false,
			allowedCodeIDs:  []uint64{12345, 42069, 69420},
			codeID:          88888,
			expAllowed:      false,
		},
	} {
		params, err := types.NewParams(tc.allowAllCodeIDs, tc.allowedCodeIDs, types.DefaultMaxGas, types.DefaultMaxGas)
		require.NoError(t, err)

		allowed := params.IsAllowed(tc.codeID)
		require.Equal(t, tc.expAllowed, allowed)
	}
}
