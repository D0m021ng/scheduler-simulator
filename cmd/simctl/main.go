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

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/templates"

	"github.com/D0m021ng/scheduler-simulator/cmd/simctl/options"
)

var logFlushFreq = pflag.Duration("log-flush-frequency", 5*time.Second,
	"maximum number of seconds between log flushes")

func main() {
	klog.InitFlags(nil)

	// The default klog flush interval is 30 seconds, which is frighteningly long.
	go wait.Until(klog.Flush, *logFlushFreq, wait.NeverStop)
	defer klog.Flush()

	rootCmd := &cobra.Command{
		Use:  "simctl",
		Long: "simctl controls test resources on simulator/kubernetes",
	}

	// tell Cobra not to provide the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	groups := templates.CommandGroups{
		{
			Message: "Basic Commands:",
			Commands: []*cobra.Command{
				options.BuildGenerateCmd(),
				options.BuildApplyCmd(),
				options.VersionCommand(),
			},
		},
		{
			Message: "Kubernetes Origin Commands:",
			Commands: []*cobra.Command{
				options.BuildKubectlCmd(),
			},
		},
	}
	groups.Add(rootCmd)
	filters := []string{"options"}
	templates.ActsAsRootCommand(rootCmd, filters, groups...)

	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("Failed to execute command: %v\n", err)
		os.Exit(2)
	}
}
