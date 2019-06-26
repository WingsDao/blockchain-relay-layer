package cli

import (
	"github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client/context"
	"wings-blockchain/x/currencies/queries"
	"github.com/cosmos/cosmos-sdk/codec"
	"fmt"
)

// Get denoms list
func GetDenoms(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "denoms",
		Short: "get denoms list",
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/denoms", queryRoute), nil)
			if err != nil {
				return err
			}

			var out queries.QueryDenomsRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}

// Get currency by denom/symbol
func GetCurrency(queryRoute string, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "currency [symbol]",
		Short: "get currency by denom/symbol",
		Args:  cobra.ExactArgs(1),
		RunE:  func(cmd *cobra.Command, args []string) error {
			cliCtx := context.NewCLIContext().WithCodec(cdc)

			res, err := cliCtx.QueryWithData(fmt.Sprintf("custom/%s/call/%s", queryRoute, args[0]), nil)
			if err != nil {
				return err
			}

			var out queries.QueryCurrencyRes
			cdc.MustUnmarshalJSON(res, &out)
			return cliCtx.PrintOutput(out)
		},
	}
}
