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
	"strings"

	"github.com/google/uuid"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	uuidMaxLen         = 32
	defaultQueue       = "default"
	queueAnnotationKey = "volcano.sh/queue-name"
)

var (
	GracePeriodSeconds int64 = 0
	PhaseList                = []v1.PodPhase{v1.PodPending, v1.PodRunning, v1.PodSucceeded, v1.PodFailed, v1.PodUnknown}
)

func BuildResources(res map[string]string) v1.ResourceList {
	var rList = make(v1.ResourceList)
	for rName, rValue := range res {
		rList[v1.ResourceName(rName)] = resource.MustParse(rValue)
	}
	return rList
}

func generateIDWithLength(Prefix string, Len int) string {
	if Len > uuidMaxLen {
		Len = uuidMaxLen
	}
	uuidStr := strings.ToLower(Prefix) + "-" + strings.ReplaceAll(uuid.NewString(), "-", "")[:Len]
	return uuidStr
}

func parseMapArgs(argsList []string) []map[string]string {
	var mapArgs []map[string]string
	if len(argsList) > 0 {
		for _, value := range argsList {
			args := strings.Split(value, ";")
			argMap := map[string]string{}
			for _, argStr := range args {
				arg := strings.Split(argStr, "=")
				if len(arg) == 2 {
					argMap[arg[0]] = arg[1]
				}
			}
			if len(argMap) > 0 {
				mapArgs = append(mapArgs, argMap)
			}
		}
	}
	return mapArgs
}

func BuildFakePod(name, namespace, schedulerName, queueName string, labels map[string]string, podPhase v1.PodPhase, req v1.ResourceList) *v1.Pod {
	if schedulerName == "" {
		schedulerName = v1.DefaultSchedulerName
	}
	if queueName == "" {
		queueName = defaultQueue
	}
	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
			Annotations: map[string]string{
				queueAnnotationKey: queueName,
			},
		},
		Spec: v1.PodSpec{
			SchedulerName:                 schedulerName,
			TerminationGracePeriodSeconds: &GracePeriodSeconds,
			Containers: []v1.Container{
				{
					Image: "nginx:latest",
					Name:  name,
					Resources: v1.ResourceRequirements{
						Requests: req,
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: podPhase,
		},
	}
}

func BuildFakeNode(name string, unsched bool, capacity, alloc v1.ResourceList, nodeConds []v1.NodeCondition, labels map[string]string) *v1.Node {
	return &v1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Labels:      labels,
			Annotations: map[string]string{},
		},
		Spec: v1.NodeSpec{
			Unschedulable: unsched,
		},
		Status: v1.NodeStatus{
			Capacity:    capacity,
			Allocatable: alloc,
			Conditions:  nodeConds,
		},
	}
}
