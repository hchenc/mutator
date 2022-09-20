/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"os"

	"github.com/hchenc/mutator/cmd/server"
	"github.com/spf13/cobra"
)

func NewMutatorCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "mutator",
		Short: "A patch tool to mutate Kubernetes resource as needed",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
	rootCmd.AddCommand(server.NewServerCommand())
	return rootCmd
}

func main() {
	cmd := NewMutatorCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
