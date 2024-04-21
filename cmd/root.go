package cmd

//goland:noinspection GoSnakeCaseUsage
import (
	"github.com/bcdevtools/chain-registry-validation-tool/cmd/dymension-chain-registry"
	"github.com/bcdevtools/chain-registry-validation-tool/constants"
	"github.com/spf13/cobra"
	"os"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.BINARY_NAME,
	Short: constants.BINARY_NAME,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	rootCmd.CompletionOptions.HiddenDefaultCmd = true    // hide the 'completion' subcommand
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true}) // hide the 'help' subcommand

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(dymension_chain_registry.Commands())

	rootCmd.PersistentFlags().Bool("help", false, "show help")
}
