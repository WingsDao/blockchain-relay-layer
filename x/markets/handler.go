package markets

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkErrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// NewHandler creates markets type messages handler.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) (*sdk.Result, error) {
		switch msg := msg.(type) {
		case MsgCreateMarket:
			return HandleMsgCreateMarket(ctx, k, msg)
		default:
			return nil, sdkErrors.Wrapf(sdkErrors.ErrUnknownRequest, "unrecognized markets message type: %T", msg)
		}
	}
}

// HandleMsgCreateMarket handles HandleMsgCreateMarket message type.
// Creates and stores new market object.
func HandleMsgCreateMarket(ctx sdk.Context, k Keeper, msg MsgCreateMarket) (*sdk.Result, error) {
	market, err := k.Add(ctx, msg.BaseAssetDenom, msg.QuoteAssetDenom)
	if err != nil {
		return nil, err
	}

	res, err := ModuleCdc.MarshalBinaryLengthPrefixed(market)
	if err != nil {
		return nil, fmt.Errorf("result marshal: %w", err)
	}

	return &sdk.Result{
		Data:   res,
		Events: ctx.EventManager().Events(),
	}, nil
}