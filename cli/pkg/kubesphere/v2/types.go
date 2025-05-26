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

package v2

type V2 struct {
	Persistence      Persistence      `yaml:"persistence"`
	Common           Common           `yaml:"common"`
	Etcd             Etcd             `yaml:"etcd"`
	MetricsServerOld MetricsServerOld `yaml:"metrics-server"`
	MetricsServerNew MetricsServerNew `yaml:"metrics_server"`
	Console          Console          `yaml:"console"`
	Monitoring       Monitoring       `yaml:"monitoring"`
	Logging          Logging          `yaml:"logging"`
	Openpitrix       Openpitrix       `yaml:"openpitrix"`
	Devops           Devops           `yaml:"devops"`
	Servicemesh      Servicemesh      `yaml:"servicemesh"`
	Notification     Notification     `yaml:"notification"`
	Alerting         Alerting         `yaml:"alerting"`
	LocalRegistry    string           `yaml:"local_registry"`
}

type Persistence struct {
	StorageClass string `yaml:"storageClass"`
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
}

type MetricsServerOld struct {
	Enabled string `yaml:"enabled"`
}

type MetricsServerNew struct {
	Enabled string `yaml:"enabled"`
}

type Console struct {
	EnableMultiLogin bool `yaml:"enableMultiLogin"`
	Port             int  `yaml:"port"`
}

type Monitoring struct {
	PrometheusReplicas      int    `yaml:"prometheusReplicas"`
	PrometheusMemoryRequest string `yaml:"prometheusMemoryRequest"`
	PrometheusVolumeSize    string `yaml:"prometheusVolumeSize"`
}

type Logging struct {
	Enabled                       bool   `yaml:"enabled"`
	ElasticsearchMasterReplicas   int    `yaml:"elasticsearchMasterReplicas"`
	ElasticsearchDataReplicas     int    `yaml:"elasticsearchDataReplicas"`
	LogsidecarReplicas            int    `yaml:"logsidecarReplicas"`
	ElasticsearchVolumeSize       string `yaml:"elasticsearchVolumeSize"`
	ElasticsearchMasterVolumeSize string `yaml:"elasticsearchMasterVolumeSize"`
	ElasticsearchDataVolumeSize   string `yaml:"elasticsearchDataVolumeSize"`
	LogMaxAge                     int    `yaml:"logMaxAge"`
	ElkPrefix                     string `yaml:"elkPrefix"`
}

type Openpitrix struct {
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

type Servicemesh struct {
	Enabled bool `yaml:"enabled"`
}

type Notification struct {
	Enabled bool `yaml:"enabled"`
}

type Alerting struct {
	Enabled bool `yaml:"enabled"`
}
