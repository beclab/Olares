// Copyright 2023 bytetrade
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

package store

import (
	"context"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kube-state-metrics/v2/pkg/metric"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
)

var (
	userAnnotationCpuLimitKey                = "bytetrade.io/user-cpu-limit"
	userAnnotationMemoryLimitKey             = "bytetrade.io/user-memory-limit"
	schemeGroupVersionResource               = schema.GroupVersionResource{Group: "iam.kubesphere.io", Version: "v1alpha2", Resource: "users"}
	descUserDefaultLabels                    = []string{"username"}
	Cpu                          CpuOrMemory = "cpu"
	Memory                       CpuOrMemory = "memory"
)

type CpuOrMemory string

func userMetricFamilies() []generator.FamilyGenerator {
	return []generator.FamilyGenerator{
		*generator.NewFamilyGenerator(
			"kube_user_cpu_total",
			"The number of user's cpu core.",
			metric.Gauge,
			"",
			wrapUserFunc(func(u *User) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   descUserDefaultLabels,
							LabelValues: []string{u.Name},
							Value:       getUserLimit(u, Cpu),
						},
					},
				}
			}),
		),
		*generator.NewFamilyGenerator(
			"kube_user_memory_total",
			"The byte of user's memory.",
			metric.Gauge,
			"",
			wrapUserFunc(func(u *User) *metric.Family {
				return &metric.Family{
					Metrics: []*metric.Metric{
						{
							LabelKeys:   descUserDefaultLabels,
							LabelValues: []string{u.Name},
							Value:       getUserLimit(u, Memory),
						},
					},
				}
			}),
		),
	}
}

func getUserLimit(u *User, metricType CpuOrMemory) float64 {
	s, _ := u.Annotations[userAnnotationCpuLimitKey]
	if metricType == Memory {
		s, _ = u.Annotations[userAnnotationMemoryLimitKey]
	}
	limit, err := resource.ParseQuantity(s)
	if err != nil {
		return float64(0)
	}

	return limit.AsApproximateFloat64()
}

func wrapUserFunc(f func(*User) *metric.Family) func(interface{}) *metric.Family {
	return func(obj interface{}) *metric.Family {
		//user := obj.(*User)
		var user User
		userUnstructured := obj.(*unstructured.Unstructured)
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(userUnstructured.UnstructuredContent(), &user)
		if err != nil {
			print("transfer from unstructured to object error", err)
		}
		metricFamily := f(&user)

		//for _, m := range metricFamily.Metrics {
		//	m.LabelKeys = append(descUserDefaultLabels, m.LabelKeys...)
		//	m.LabelValues = append([]string{user.Name}, m.LabelValues...)
		//}
		return metricFamily
	}
}

func createUserListWatchFunc(userClient dynamic.Interface) func(kubeClient clientset.Interface, _ string, _ string) cache.ListerWatcher {
	return func(kubeClient clientset.Interface, _ string, _ string) cache.ListerWatcher {
		return &cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				return userClient.Resource(schemeGroupVersionResource).List(context.TODO(), opts)
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				return userClient.Resource(schemeGroupVersionResource).Watch(context.TODO(), opts)
			},
		}
	}
}
