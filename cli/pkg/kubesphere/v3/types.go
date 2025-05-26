/*
 Copyright 2021 The KubeSphere Authors.

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

package v3

type ClusterConfig struct {
	ApiVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       *V3      `yaml:"spec"`
}

type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Label     Label  `yaml:"labels"`
}

type Label struct {
	Version string `yaml:"version"`
}

type V3 struct {
	Persistence    Persistence    `yaml:"persistence"`
	Authentication Authentication `yaml:"authentication"`
	Common         Common         `yaml:"common"`
	Etcd           Etcd           `yaml:"etcd"`
	MetricsServer  MetricsServer  `yaml:"metrics_server"`
	Console        Console        `yaml:"console"`
	Monitoring     Monitoring     `yaml:"monitoring"`
	Logging        Logging        `yaml:"logging"`
	Openpitrix     Openpitrix     `yaml:"openpitrix"`
	Devops         Devops         `yaml:"devops"`
	Servicemesh    Servicemesh    `yaml:"servicemesh"`
	Notification   Notification   `yaml:"notification"`
	Alerting       Alerting       `yaml:"alerting"`
	Auditing       Auditing       `yaml:"auditing"`
	Events         Events         `yaml:"events"`
	Multicluster   Multicluster   `yaml:"multicluster"`
	Networkpolicy  Networkpolicy  `yaml:"networkpolicy"`
	LocalRegistry  string         `yaml:"local_registry"`
}

type Persistence struct {
	StorageClass string `yaml:"storageClass"`
}

type MetricsServer struct {
	Enabled bool `yaml:"enabled"`
}

type Authentication struct {
	JwtSecret string `yaml:"jwtSecret"`
}
type Etcd struct {
	Monitoring  bool   `yaml:"monitoring"`
	EndpointIps string `yaml:"endpointIps"`
	Port        int    `yaml:"port"`
	TlsEnable   bool   `yaml:"tlsEnable"`
}

type Common struct {
	MysqlVolumeSize    string `yaml:"mysqlVolumeSize"`
	MinioVolumeSize    string `yaml:"minioVolumeSize"`
	EtcdVolumeSize     string `yaml:"etcdVolumeSize"`
	OpenldapVolumeSize string `yaml:"openldapVolumeSize"`
	RedisVolumSize     string `yaml:"redisVolumSize"`
	ES                 ES     `yaml:"es"`
}

type ES struct {
	//ElasticsearchMasterReplicas   int    `yaml:"elasticsearchMasterReplicas"`
	//ElasticsearchDataReplicas     int    `yaml:"elasticsearchDataReplicas"`
	ElasticsearchMasterVolumeSize string `yaml:"elasticsearchMasterVolumeSize"`
	ElasticsearchDataVolumeSize   string `yaml:"elasticsearchDataVolumeSize"`
	LogMaxAge                     int    `yaml:"logMaxAge"`
	ElkPrefix                     string `yaml:"elkPrefix"`
}

type Console struct {
	EnableMultiLogin bool `yaml:"enableMultiLogin"`
	Port             int  `yaml:"port"`
}

type Alerting struct {
	Enabled bool `yaml:"enabled"`
}

type Auditing struct {
	Enabled bool `yaml:"enabled"`
}

type Devops struct {
	Enabled               bool   `yaml:"enabled"`
	JenkinsMemoryLim      string `yaml:"jenkinsMemoryLim"`
	JenkinsMemoryReq      string `yaml:"jenkinsMemoryReq"`
	JenkinsVolumeSize     string `yaml:"jenkinsVolumeSize"`
	JenkinsjavaoptsXms    string `yaml:"jenkinsJavaOpts_Xms"`
	JenkinsjavaoptsXmx    string `yaml:"jenkinsJavaOpts_Xmx"`
	JenkinsjavaoptsMaxram string `yaml:"jenkinsJavaOpts_MaxRAM"`
}

type Events struct {
	Enabled bool  `yaml:"enabled"`
	Ruler   Ruler `yaml:"ruler"`
}

type Ruler struct {
	Enabled  bool `yaml:"enabled"`
	Replicas int  `yaml:"replicas"`
}

type Logging struct {
	Enabled            bool `yaml:"enabled"`
	LogsidecarReplicas int  `yaml:"logsidecarReplicas"`
}

type Metrics struct {
	Enabled bool `yaml:"enabled"`
}

type Monitoring struct {
	//AlertmanagerReplicas    int    `yaml:"alertmanagerReplicas"`
	//PrometheusReplicas      int    `yaml:"prometheusReplicas"`
	PrometheusMemoryRequest string `yaml:"prometheusMemoryRequest"`
	PrometheusVolumeSize    string `yaml:"prometheusVolumeSize"`
}

type Multicluster struct {
	ClusterRole string `yaml:"clusterRole"`
}

type Networkpolicy struct {
	Enabled bool `yaml:"enabled"`
}

type Notification struct {
	Enabled bool `yaml:"enabled"`
}

type Openpitrix struct {
	Enabled bool `yaml:"enabled"`
}

type Servicemesh struct {
	Enabled bool `yaml:"enabled"`
}
