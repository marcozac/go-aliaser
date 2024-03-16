package internal

import "github.com/spf13/cobra"

func NewRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "aliaser",
		Short: "aliaser is a tool to generate aliases from a Go package",
	}
	cmd.AddCommand(NewGenerate())
	return cmd
}
