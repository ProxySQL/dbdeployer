package cmd

import (
	"fmt"

	"github.com/ProxySQL/dbdeployer/providers"
	"github.com/spf13/cobra"
)

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "Shows available deployment providers",
	Long:  "Lists all registered providers that can be used for sandbox deployment",
	Run: func(cmd *cobra.Command, args []string) {
		for _, name := range providers.DefaultRegistry.List() {
			p, _ := providers.DefaultRegistry.Get(name)
			ports := p.DefaultPorts()
			fmt.Printf("%-15s (base port: %d, ports per instance: %d)\n",
				name, ports.BasePort, ports.PortsPerInstance)
		}
	},
}

func init() {
	rootCmd.AddCommand(providersCmd)
}
