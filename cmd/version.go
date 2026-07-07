package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionShort bool

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the acolyte binary version",
	Long: `Print the version of the acolyte binary.

The version is embedded at build time via -ldflags '-X main.Version=vX.Y.Z'.
Falls back to 'dev' when built from a local checkout without an injected version.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if versionShort {
			fmt.Println(versionValue)
			return nil
		}
		fmt.Printf("acolyte %s\n", versionValue)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVar(&versionShort, "short", false, "Print only the version string (e.g. v1.2.3)")
	rootCmd.AddCommand(versionCmd)
}
