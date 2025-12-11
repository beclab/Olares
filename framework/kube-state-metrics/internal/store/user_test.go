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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	generator "k8s.io/kube-state-metrics/v2/pkg/metric_generator"
	"testing"
)

func TestUserStore(t *testing.T) {
	//const metadata = `
	//    # HELP kube_user_cpu_total The number of user's cpu core.
	//	# TYPE kube_user_cpu_total gauge
	//`
	const metadata = `
	    # HELP kube_user_cpu_total The number of user's cpu core.
		# TYPE kube_user_cpu_total gauge
	    # HELP kube_user_memory_total The byte of user's memory.
		# TYPE kube_user_memory_total gauge
	`
	cases := []generateMetricsTestCase{
		{
			AllowAnnotationsList: []string{userAnnotationCpuLimitKey, userAnnotationMemoryLimitKey},
			Obj: &User{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "bob",
					Namespace: "ns1",
					Annotations: map[string]string{
						userAnnotationCpuLimitKey:    "1",
						userAnnotationMemoryLimitKey: "1G",
					},
				},
			},
			Want: metadata + `
        kube_user_cpu_total{user="bob"} 1
		kube_user_memory_total{user="bob"} 1e+09
       `,
		},
	}
	for i, c := range cases {
		c.Func = generator.ComposeMetricGenFuncs(userMetricFamilies())
		c.Headers = generator.ExtractMetricFamilyHeaders(userMetricFamilies())
		if err := c.run(); err != nil {
			t.Errorf("unexpected collecting result in %vth run:\n%s", i, err)
		}
	}
}
