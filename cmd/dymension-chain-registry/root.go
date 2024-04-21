package dymension_chain_registry

import (
	"github.com/spf13/cobra"
)

// Commands registers a sub-tree of commands
func Commands() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dymension-chain-registry",
		Aliases: []string{"dcr", "dym"},
		Short:   "Tools for Dymension chain-registry at https://github.com/dymensionxyz/chain-registry",
	}

	cmd.AddCommand(
		GetValidateCommand(),
	)

	return cmd
}
