package internal

import (
	"fmt"

	"github.com/marcozac/go-aliaser"
	"github.com/spf13/cobra"
)

func init() {
	generateCmd.Flags().String("pattern", "", "the package pattern, in go format, to generate aliases for")
	generateCmd.Flags().String("target", "", "the package name to use in the generated file")
	generateCmd.Flags().String("file", "", "the file name to write the aliases to")
	generateCmd.Flags().String("header", "", "optional header to be written at the top of the file")
	generateCmd.Flags().Bool("exclude-constants", false, "exclude constants from the generated aliases")
	generateCmd.Flags().Bool("exclude-variables", false, "exclude variables from the generated aliases")
	generateCmd.Flags().Bool("exclude-functions", false, "exclude functions from the generated aliases")
	generateCmd.Flags().Bool("exclude-types", false, "exclude types from the generated aliases")
	generateCmd.Flags().StringSlice("exclude-names", nil, "exclude specific names from the generated aliases")
	generateCmd.Flags().Bool("dry-run", false, "print the aliases without writing them to the file")

	Must(generateCmd.MarkFlagRequired("pattern"))
	Must(generateCmd.MarkFlagRequired("target"))
	generateCmd.MarkFlagsOneRequired("file", "dry-run")
}

var generateCmd = &cobra.Command{
	Use: "generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		opts := []aliaser.Option{aliaser.WithContext(cmd.Context())}
		if MustV(cmd.Flags().GetBool("exclude-constants")) {
			opts = append(opts, aliaser.ExcludeConstants())
		}
		if MustV(cmd.Flags().GetBool("exclude-variables")) {
			opts = append(opts, aliaser.ExcludeVariables())
		}
		if MustV(cmd.Flags().GetBool("exclude-functions")) {
			opts = append(opts, aliaser.ExcludeFunctions())
		}
		if MustV(cmd.Flags().GetBool("exclude-types")) {
			opts = append(opts, aliaser.ExcludeTypes())
		}
		if names := MustV(cmd.Flags().GetStringSlice("exclude-names")); len(names) > 0 {
			opts = append(opts, aliaser.ExcludeNames(names...))
		}
		if header := MustV(cmd.Flags().GetString("header")); header != "" {
			opts = append(opts, aliaser.WithHeader(header))
		}
		a, err := aliaser.New(
			MustV(cmd.Flags().GetString("target")),
			MustV(cmd.Flags().GetString("pattern")),
			opts...,
		)
		if err != nil {
			return fmt.Errorf("aliaser: %w", err)
		}
		if MustV(cmd.Flags().GetBool("dry-run")) {
			return a.Generate(cmd.OutOrStdout())
		}
		return a.GenerateFile(MustV(cmd.Flags().GetString("file")))
	},
}
