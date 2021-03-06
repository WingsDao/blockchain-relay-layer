package rest

import (
	"github.com/cosmos/cosmos-sdk/x/auth"

	"github.com/dfinance/dnode/x/multisig/internal/types"
)

//nolint:deadcode,unused
type (
	MSRespGetCall struct {
		Height int64          `json:"height"`
		Result types.CallResp `json:"result"`
	}

	MSRespGetCalls struct {
		Height int64           `json:"height"`
		Result types.CallsResp `json:"result"`
	}

	CCRespStdTx struct {
		Height int64      `json:"height"`
		Result auth.StdTx `json:"result"`
	}
)
