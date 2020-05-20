package run

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the network-qos version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
}
