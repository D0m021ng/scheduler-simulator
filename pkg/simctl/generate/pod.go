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
	podNSList    = []string{"default"}
	podQueueList = []string{"default"}
	podPhaseList = []v1.PodPhase{v1.PodPending}
	podReqList   = []map[string]string{
		{
			"cpu":    "2",
			"memory": "4Gi",
		},
		{
			"cpu":    "4",
			"memory": "8Gi",
		},
	}
	podLabelsList = []map[string]string{
		{
			"scheduler-simulator": "true",
		},
	}
)

type generatePodFlags struct {
	Count         int
	ResourceList  []string
	QueueList     []string
	NamespaceList []string
	LabelList     []string

	Output        string
	SchedulerName string
}

var genPodFlags = &generatePodFlags{}

// InitGeneratePodFlags is used to init all flags during generate pod data.
func InitGeneratePodFlags(cmd *cobra.Command) {

	cmd.Flags().IntVarP(&genPodFlags.Count, "count", "c", 1, "the count of pods")
	cmd.Flags().StringVarP(&genPodFlags.Output, "output", "o", "testdata-pod.yaml", "the name of pod test data file")
	cmd.Flags().StringVarP(&genPodFlags.SchedulerName, "schedulerName", "n", "volcano", "the name of scheduler")
	cmd.Flags().StringSliceVarP(&genPodFlags.QueueList, "queues", "q", []string{"default"}, "queues for pods")
	cmd.Flags().StringSliceVarP(&genPodFlags.NamespaceList, "namespaces", "", []string{"default"}, "namespaces for pods")
	// Map arguments
	// For example:
	//   -r "cpu=2;memory=4Gi" -r "cpu=4;memory=8Gi;nvidia.com/gpu=1"
	// will result in
	//   []map[string]string{
	//      { "cpu": "2", "memory": "4Gi"},
	//      { "cpu": "4", "memory": "8Gi", "nvidia.com/gpu": "1"},
	//    }
	cmd.Flags().StringSliceVarP(&genPodFlags.ResourceList, "resources", "r",
		nil, "the resource list for pods. e.g. -r \"cpu=2;memory=4Gi\" -r \"cpu=4;memory=8Gi;nvidia.com/gpu=1\" ")
	cmd.Flags().StringSliceVarP(&genPodFlags.LabelList, "labels", "l",
		nil, "labels for pods. e.g. --labels \"a=b\" --labels \"a=d;c=e\" ")
}

func GenFakePods() error {
	if len(genPodFlags.QueueList) > 0 {
		podQueueList = genPodFlags.QueueList
	}
	if len(genPodFlags.NamespaceList) > 0 {
		podNSList = genPodFlags.NamespaceList
	}
	if len(genPodFlags.ResourceList) != 0 {
		podReqList = parseMapArgs(genPodFlags.ResourceList)
	}
	if len(genPodFlags.LabelList) != 0 {
		podLabelsList = parseMapArgs(genPodFlags.LabelList)
	}
	fmt.Printf("Generate test data of %d pod(s) with following config: \n", genPodFlags.Count)
	fmt.Printf("Pod namespace list: %s\n", podNSList)
	fmt.Printf("Pod queue list: %s\n", podQueueList)
	fmt.Printf("Pod request resources list: %s\n", podReqList)
	fmt.Printf("Pod labels list: %s\n", podLabelsList)

	return fakePods(genPodFlags.Count, podNSList, podQueueList, podPhaseList, podReqList, podLabelsList)
}

func fakePods(podCount int, nsList, queueList []string, phaseList []v1.PodPhase, reqList, labelsList []map[string]string) error {
	nsLen := len(nsList)
	queueLen := len(queueList)
	phaseLen := len(phaseList)
	reqLen := len(reqList)
	labelsLen := len(labelsList)

	var name, namespace string
	var podsYaml []byte
	var err error
	rand.Seed(time.Now().Unix())

	for idx := 0; idx < podCount; idx++ {
		name = generateIDWithLength("test-pod", 16)
		namespace = nsList[rand.Intn(nsLen)]
		reqRes := reqList[rand.Intn(reqLen)]
		queueName := queueList[rand.Intn(queueLen)]
		// generate labels for pod
		labels := labelsList[rand.Intn(labelsLen)]

		fakePod := BuildFakePod(name, namespace, genPodFlags.SchedulerName, queueName, labels, phaseList[rand.Intn(phaseLen)], BuildResources(reqRes))
		fakePodStr, err := yaml.Marshal(fakePod)
		if err != nil {
			fmt.Printf("json marshal failed, err: %v", err)
			break
		}
		fakePodStr = append(fakePodStr, []byte("---\n")...)
		podsYaml = append(podsYaml, fakePodStr...)
	}

	// write test data to file
	yamlfile, err := os.OpenFile(genPodFlags.Output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("error opening/creating file: %v", err)
		return err
	}
	_, err = yamlfile.Write(podsYaml)
	if err != nil {
		fmt.Printf("json marshal failed, err: %v", err)
		return err
	}
	return err
}
