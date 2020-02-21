// Genesis block related commands for VM module.
package cli

import (
	"encoding/json"
	"github.com/WingsDao/wings-blockchain/x/vm/internal/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/libs/cli"
	tmdTypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"os"
)

// Reading genesis state from file generated by Move VM.
// File contains write set operations for standard libraries.
func GenesisWSFromFile(ctx *server.Context, cdc *codec.Codec) *cobra.Command {
	return &cobra.Command{
		Use:   "read-genesis-write-set [writeSetJsonFile]",
		Short: "Read write set from json file and place into genesis state, if write set already exists - will be rewritten",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			config := ctx.Config
			config.SetRoot(viper.GetString(cli.HomeFlag))

			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()

			jsonContent, err := ioutil.ReadAll(file)
			if err != nil {
				return err
			}

			var genesisState types.GenesisState
			if err := json.Unmarshal(jsonContent, &genesisState); err != nil {
				return err
			}

			genFile := config.GenesisFile()
			genDoc, err := tmdTypes.GenesisDocFromFile(genFile)
			if err != nil {
				return err
			}

			appState, err := genutil.GenesisStateFromGenDoc(cdc, *genDoc)
			if err != nil {
				return err
			}

			genesisStateBz := cdc.MustMarshalJSON(genesisState)
			appState[types.ModuleName] = genesisStateBz

			appStateJson, err := cdc.MarshalJSON(appState)
			if err != nil {
				return err
			}

			genDoc.AppState = appStateJson
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}
}
