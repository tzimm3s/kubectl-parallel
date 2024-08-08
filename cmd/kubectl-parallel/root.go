package main

import (
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	var params rootCommandParams

	cmd := &cobra.Command{
		Use: "kubectl-parallel",
	}

	cmd.PersistentFlags().StringVarP(&params.label, "label", "l", defaultLabel, "label for grouping resources")

	cmd.AddCommand(
		NewApplyCommand(&params),
	)

	return cmd
}

type rootCommandParams struct {
	label     string
	namespace string
	verbosity int
}
