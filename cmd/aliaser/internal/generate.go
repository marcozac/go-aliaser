package internal

import (
	"github.com/marcozac/go-aliaser"
	"github.com/spf13/cobra"
)

func init() {
	generateCmd.Flags().String("from", "", "the package path, in Go format, to generate aliases from")
	generateCmd.Flags().String("package", "", "the package name to use in the generated file")
	generateCmd.Flags().String("file", "", "the file name to write the aliases to")
	generateCmd.Flags().String("header", "", "an optional header to be written at the top of the file")
	generateCmd.Flags().Bool("exclude-constants", false, "exclude constants from the generated aliases")
	generateCmd.Flags().Bool("exclude-variables", false, "exclude variables from the generated aliases")
	generateCmd.Flags().Bool("exclude-functions", false, "exclude functions from the generated aliases")
	generateCmd.Flags().Bool("exclude-types", false, "exclude types from the generated aliases")
	generateCmd.Flags().StringSlice("exclude-names", nil, "exclude specific names from the generated aliases")
	generateCmd.Flags().Bool("dry-run", false, "print the aliases without writing them to the file")

	Must(generateCmd.MarkFlagRequired("from"))
	Must(generateCmd.MarkFlagRequired("package"))
	generateCmd.MarkFlagsOneRequired("file", "dry-run")
}

var generateCmd = &cobra.Command{
	Use: "generate",
	RunE: func(cmd *cobra.Command, args []string) error {
		a := &aliaser.Alias{
			PkgName: MustV(cmd.Flags().GetString("package")),
			Header:  MustV(cmd.Flags().GetString("header")),
		}
		opts := []aliaser.Option{aliaser.WithContext(cmd.Context())}
		if MustV(cmd.Flags().GetBool("dry-run")) {
			opts = append(opts, aliaser.WithWriter(cmd.OutOrStdout()))
		} else {
			a.Out = MustV(cmd.Flags().GetString("file"))
		}
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
		src, err := aliaser.Load(MustV(cmd.Flags().GetString("from")), opts...)
		if err != nil {
			return err
		}
		a.Src = src
		return aliaser.Generate(a, opts...)
	},
}
