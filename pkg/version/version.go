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

package version

import (
	"fmt"
	"runtime"
)

var (
	GitVersion = "v0.0.0"
	GitCommit  = "unknown"
	GitBranch  = "unknown"
	BuildDate  = "unknown"
)

func info() []string {
	return []string{
		fmt.Sprintf("GitVersion: %v", GitVersion),
		fmt.Sprintf("GitCommit: %v", GitCommit),
		fmt.Sprintf("GitBranch: %v", GitBranch),
		fmt.Sprintf("BuildDate: %v", BuildDate),
		fmt.Sprintf("GoVersion: %v", runtime.Version()),
		fmt.Sprintf("Compiler: %v", runtime.Compiler),
		fmt.Sprintf("Platform: %s", fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)),
	}
}

func PrintVersionAndExit() {
	for _, i := range info() {
		fmt.Println(i)
	}
}
