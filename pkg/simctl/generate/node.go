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

package generate

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

var (
	nodeConditions = []v1.NodeCondition{
		{
			Message: "kubelet is posting ready status",
			Reason:  "KubeletReady",
			Status:  "True",
			Type:    "Ready",
		},
	}
	nodeLabels = []map[string]string{
		{
			"kubernetes.io/arch":           "amd64",
			"kubernetes.io/os":             "linux",
			"node-role.kubernetes.io/node": "",
		},
	}
	nodeResources = []map[string]string{
		{
			"cpu":    "24",
			"memory": "128Gi",
		},
		{
			"cpu":            "48",
			"memory":         "256Gi",
			"nvidia.com/gpu": "8",
		},
	}
)

type generateNodeFlags struct {
	Output        string
	Count         int
	ResourcesList []string
	LabelsList    []string
}

var genNodeFlags = &generateNodeFlags{}

// InitGenerateNodeFlags is used to init all flags during generate data.
func InitGenerateNodeFlags(cmd *cobra.Command) {

	cmd.Flags().IntVarP(&genNodeFlags.Count, "count", "c", 1, "the count of nodes")
	cmd.Flags().StringVarP(&genNodeFlags.Output, "output", "o", "testdata-node.yaml", "the name of node test data file")
	cmd.Flags().StringSliceVarP(&genNodeFlags.ResourcesList, "resources", "r",
		nil, "the resources list for nodes. e.g. -r \"cpu=24;memory=128Gi\" -r \"cpu=48;memory=128Gi;nvidia.com/gpu=8\" ")
	cmd.Flags().StringSliceVarP(&genNodeFlags.LabelsList, "labels", "l",
		nil, "the labels for nodes. e.g. --labels \"a=b\" -r \"a=c;d=b\" ")
}

func GenFakeNode() error {
	if len(genNodeFlags.ResourcesList) > 0 {
		nodeResources = parseMapArgs(genNodeFlags.ResourcesList)
	}
	if len(genNodeFlags.LabelsList) > 0 {
		nodeLabels = parseMapArgs(genNodeFlags.LabelsList)
	}
	fmt.Printf("Generate test data of %d node(s) with following config: \n", genNodeFlags.Count)
	fmt.Printf("Node capacity resources list: %s\n", nodeResources)
	fmt.Printf("Node labels list: %s\n", nodeLabels)
	return fakeNodes(genNodeFlags.Count, nodeResources, nodeLabels)
}

func fakeNodes(nodeCount int, resourceList []map[string]string, labelList []map[string]string) error {
	allocLen := len(resourceList)
	labelLen := len(labelList)

	var name string
	var nodesYaml []byte
	rand.Seed(time.Now().Unix())
	for idx := 1; idx <= nodeCount; idx++ {
		name = fmt.Sprintf("instance-%04d", idx)
		// generate node labels
		labels := labelList[rand.Intn(labelLen)]
		if labels == nil {
			labels = map[string]string{}
		}
		labels["kubernetes.io/hostname"] = name
		// generate node resources
		nodeRes := resourceList[rand.Intn(allocLen)]
		nodeRes["pods"] = "110"
		capacity, alloc := genNodeResources(BuildResources(nodeRes))
		fakeNode := BuildFakeNode(name, false, capacity, alloc, nodeConditions, labels)
		nodeStr, err := yaml.Marshal(fakeNode)
		if err != nil {
			fmt.Printf("json marshal failed, err: %v", err)
			break
		}
		nodeStr = append(nodeStr, []byte("---\n")...)
		nodesYaml = append(nodesYaml, nodeStr...)
	}

	// write test data to file
	yamlfile, err := os.OpenFile(genNodeFlags.Output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("error opening/creating file: %v", err)
		return err
	}
	_, err = yamlfile.Write(nodesYaml)
	if err != nil {
		fmt.Printf("json marshal failed, err: %v", err)
		return err
	}
	return nil
}

func genNodeResources(res v1.ResourceList) (v1.ResourceList, v1.ResourceList) {
	// TODO: reserve resource for node
	return res, res
}
