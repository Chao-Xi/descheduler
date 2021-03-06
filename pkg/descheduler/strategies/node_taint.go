/*
Copyright 2017 The Kubernetes Authors.

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

package strategies

import (
	"context"

	"sigs.k8s.io/descheduler/pkg/api"
	"sigs.k8s.io/descheduler/pkg/descheduler/evictions"
	podutil "sigs.k8s.io/descheduler/pkg/descheduler/pod"
	"sigs.k8s.io/descheduler/pkg/utils"

	v1 "k8s.io/api/core/v1"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// RemovePodsViolatingNodeTaints evicts pods on the node which violate NoSchedule Taints on nodes
func RemovePodsViolatingNodeTaints(ctx context.Context, client clientset.Interface, strategy api.DeschedulerStrategy, nodes []*v1.Node, podEvictor *evictions.PodEvictor) {
	for _, node := range nodes {
		klog.V(1).Infof("Processing node: %#v\n", node.Name)
		pods, err := podutil.ListPodsOnANode(ctx, client, node, podEvictor.IsEvictable)
		if err != nil {
			//no pods evicted as error encountered retrieving evictable Pods
			return
		}
		totalPods := len(pods)
		for i := 0; i < totalPods; i++ {
			if !utils.TolerationsTolerateTaintsWithFilter(
				pods[i].Spec.Tolerations,
				node.Spec.Taints,
				func(taint *v1.Taint) bool { return taint.Effect == v1.TaintEffectNoSchedule },
			) {
				klog.V(2).Infof("Not all taints with NoSchedule effect are tolerated after update for pod %v on node %v", pods[i].Name, node.Name)
				if _, err := podEvictor.EvictPod(ctx, pods[i], node, "NodeTaint"); err != nil {
					klog.Errorf("Error evicting pod: (%#v)", err)
					break
				}
			}
		}
	}
}
