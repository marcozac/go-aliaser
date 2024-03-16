package internal

import "github.com/spf13/cobra"

func init() {
	Root.AddCommand(generateCmd)
}

var Root = &cobra.Command{
	Use:   "aliaser",
	Short: "aliaser is a tool to generate aliases from a Go package",
}
