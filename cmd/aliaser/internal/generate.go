package internal

import (
	"fmt"

	"github.com/marcozac/go-aliaser"
	"github.com/spf13/cobra"
)

func NewGenerate() *cobra.Command {
	cmd := &cobra.Command{
		Use: "generate",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := []aliaser.Option{
				aliaser.WithContext(cmd.Context()),
				aliaser.WithPatterns(MustV(cmd.Flags().GetStringSlice("patterns"))...),
				aliaser.ExcludeConstants(MustV(cmd.Flags().GetBool("exclude-constants"))),
				aliaser.ExcludeVariables(MustV(cmd.Flags().GetBool("exclude-variables"))),
				aliaser.ExcludeFunctions(MustV(cmd.Flags().GetBool("exclude-functions"))),
				aliaser.ExcludeTypes(MustV(cmd.Flags().GetBool("exclude-types"))),
				aliaser.ExcludeNames(MustV(cmd.Flags().GetStringSlice("exclude-names"))...),
				aliaser.AssignFunctions(MustV(cmd.Flags().GetBool("assign-functions"))),
			}
			if header := MustV(cmd.Flags().GetString("header")); header != "" {
				opts = append(opts, aliaser.WithHeader(header))
			}
			a, err := aliaser.New(aliaser.Config{
				TargetPackage: MustV(cmd.Flags().GetString("target")),
			}, opts...)
			if err != nil {
				return fmt.Errorf("aliaser: %w", err)
			}
			if MustV(cmd.Flags().GetBool("dry-run")) {
				return a.Generate(cmd.OutOrStdout())
			}
			return a.GenerateFile(MustV(cmd.Flags().GetString("file")))
		},
	}
	cmd.Flags().String("target", "", "the package name to use in the generated file")
	cmd.Flags().StringSlice("patterns", []string{}, "the package patterns, in go format, to generate aliases for")
	cmd.Flags().String("file", "", "the file name to write the aliases to")
	cmd.Flags().String("header", "", "optional header to be written at the top of the file")
	cmd.Flags().Bool("exclude-constants", false, "exclude constants from the generated aliases")
	cmd.Flags().Bool("exclude-variables", false, "exclude variables from the generated aliases")
	cmd.Flags().Bool("exclude-functions", false, "exclude functions from the generated aliases")
	cmd.Flags().Bool("exclude-types", false, "exclude types from the generated aliases")
	cmd.Flags().StringSlice("exclude-names", nil, "exclude specific names from the generated aliases")
	cmd.Flags().Bool("assign-functions", false, "assign functions to variables in the generated aliases")
	cmd.Flags().Bool("dry-run", false, "print the aliases without writing them to the file")

	Must(cmd.MarkFlagRequired("target"))
	cmd.MarkFlagsOneRequired("file", "dry-run")
	return cmd
}
