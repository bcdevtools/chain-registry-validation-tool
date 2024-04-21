package dymension_chain_registry

import (
	"github.com/spf13/cobra"
)

func GetValidateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate [repo-dir]",
		Short: "Validate Dymension chain-registry",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	return cmd
}
