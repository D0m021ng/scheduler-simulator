/*
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

package options

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/kubectl/pkg/cmd"

	"github.com/D0m021ng/scheduler-simulator/pkg/simctl/apply"
	"github.com/D0m021ng/scheduler-simulator/pkg/simctl/generate"
	"github.com/D0m021ng/scheduler-simulator/pkg/version"
)

func checkError(cmd *cobra.Command, err error) {
	if err != nil {
		msg := "Failed to"

		// Ignore the root command.
		for cur := cmd; cur.Parent() != nil; cur = cur.Parent() {
			msg += fmt.Sprintf(" %s", cur.Name())
		}

		fmt.Printf("%s: %v\n", msg, err)
		os.Exit(2)
	}
}

func BuildGenerateCmd() *cobra.Command {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate fake test data",
	}

	genNodeCmd := &cobra.Command{
		Use:   "node",
		Short: "Generate fake node data for testing",
		Run: func(cmd *cobra.Command, args []string) {
			checkError(cmd, generate.GenFakeNode())
		},
	}
	generate.InitGenerateNodeFlags(genNodeCmd)
	generateCmd.AddCommand(genNodeCmd)

	genPodCmd := &cobra.Command{
		Use:   "pod",
		Short: "Generate fake pod data for testing",
		Run: func(cmd *cobra.Command, args []string) {
			checkError(cmd, generate.GenFakePods())
		},
	}
	generate.InitGeneratePodFlags(genPodCmd)
	generateCmd.AddCommand(genPodCmd)

	return generateCmd
}

func BuildApplyCmd() *cobra.Command {
	return apply.NewCmdApply(genericclioptions.IOStreams{In: os.Stdin, Out: os.Stdout, ErrOut: os.Stderr})
}

func BuildKubectlCmd() *cobra.Command {
	return cmd.NewDefaultKubectlCommand()
}

func VersionCommand() *cobra.Command {
	var command = &cobra.Command{
		Use:     "version",
		Short:   "Print the version information",
		Long:    "Print the version information",
		Example: "simctl version",
		Run: func(cmd *cobra.Command, args []string) {
			version.PrintVersionAndExit()
		},
	}
	return command
}
