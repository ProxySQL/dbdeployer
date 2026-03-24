// DBDeployer - The MySQL Sandbox
// Copyright © 2006-2021 Giuseppe Maxia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
