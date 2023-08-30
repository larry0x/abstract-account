package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/larry0x/abstract-account/x/abstractaccount/types"
)

func TestValidateParams(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		params *types.Params
		expErr bool
	}{
		{
			desc:   "max gas is nil",
			params: &types.Params{},
			expErr: true,
		},
		{
			desc:   "max gas before is zero",
			params: &types.Params{MaxGasBefore: 0, MaxGasAfter: types.DefaultMaxGas},
			expErr: true,
		},
		{
			desc:   "max gas after is nil",
			params: &types.Params{MaxGasBefore: types.DefaultMaxGas, MaxGasAfter: 0},
			expErr: true,
		},
		{
			desc: "allow list is not empty when AllowAllCodeIDs is true",
			params: &types.Params{
				AllowAllCodeIDs: true,
				AllowedCodeIDs:  []uint64{1, 2, 3},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: true,
		},
		{
			desc: "allow list contains zero code ID",
			params: &types.Params{
				AllowAllCodeIDs: false,
				AllowedCodeIDs:  []uint64{1, 0, 3},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: true,
		},
		{
			desc: "allow list contains duplicate code IDs",
			params: &types.Params{
				AllowAllCodeIDs: false,
				AllowedCodeIDs:  []uint64{1, 2, 2, 3},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: true,
		},
		{
			desc: "allow list contains unsorted code IDs",
			params: &types.Params{
				AllowAllCodeIDs: false,
				AllowedCodeIDs:  []uint64{1, 2, 3, 2},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: true,
		},
		{
			desc: "valid params - all code IDs allowed",
			params: &types.Params{
				AllowAllCodeIDs: true,
				AllowedCodeIDs:  []uint64{},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: false,
		},
		{
			desc: "valid params - not all code IDs allowed",
			params: &types.Params{
				AllowAllCodeIDs: false,
				AllowedCodeIDs:  []uint64{1, 2, 3},
				MaxGasBefore:    types.DefaultMaxGas,
				MaxGasAfter:     types.DefaultMaxGas,
			},
			expErr: false,
		},
	} {
		err := tc.params.Validate()

		if tc.expErr {
			require.Error(t, err, tc.desc)
		} else {
			require.NoError(t, err, tc.desc)
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
