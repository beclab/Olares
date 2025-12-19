/*
 * This file is part of KubeSphere Console.
 * Copyright (C) 2019 The KubeSphere Console Authors.
 *
 * KubeSphere Console is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * KubeSphere Console is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with KubeSphere Console.  If not, see <https://www.gnu.org/licenses/>.
 */

const TencentKubernetesEngine = {
	'service.kubernetes.io/loadbalance-id': '',
	'service.kubernetes.io/qcloud-loadbalancer-internal-subnetid': '',
	'service.kubernetes.io/tke-existed-lbid': '',
	'service.kubernetes.io/local-svc-only-bind-node-with-pod': '',
	'service.cloud.tencent.com/local-svc-weighted-balance': '',
	'service.kubernetes.io/qcloud-loadbalancer-backends-label': '',
	'service.cloud.tencent.com/direct-access': '',
	'service.cloud.tencent.com/tke-service-config': '',
	'service.cloud.tencent.com/tke-service-config-auto': '',
	'service.kubernetes.io/loadbalance-nat-ipv6': '',
	'service.kubernetes.io/loadbalance-type': '',
	'service.cloud.tencent.com/specify-protocol': '',
	'service.kubernetes.io/service.extensiveParameters': '',
	'service.cloud.tencent.com/enable-grace-shutdown': '',
	'kubernetes.io/service.internetChargeType': '',
	'kubernetes.io/service.internetMaxBandwidthOut': ''
};

const QingCloud = {
	'service.beta.kubernetes.io/qingcloud-load-balancer-eip-ids': '',
	'service.beta.kubernetes.io/qingcloud-load-balancer-type': '0'
};

const Huawei = {
	'kubernetes.io/elb.class': '',
	'kubernetes.io/session-affinity-mode': '',
	'kubernetes.io/elb.id': '',
	'kubernetes.io/elb.subnet-id': ''
};
const GoogleKubernetesEngine = {
	'cloud.google.com/network-tier': ''
};

const microsoft = {
	'service.beta.kubernetes.io/azure-load-balancer-internal': '',
	'service.beta.kubernetes.io/azure-load-balancer-internal-subnet': '',
	'service.beta.kubernetes.io/azure-dns-label-name': '',
	'service.beta.kubernetes.io/azure-shared-securityrule': '',
	'service.beta.kubernetes.io/azure-load-balancer-resource-group': '',
	'service.beta.kubernetes.io/azure-allowed-service-tags': '',
	'service.beta.kubernetes.io/azure-load-balancer-tcp-idle-timeout': '',
	'service.beta.kubernetes.io/azure-load-balancer-disable-tcp-reset': ''
};

const aliyun = {
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-address-type': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-charge-type': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-id': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-spec	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-master-zoneid	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-slave-zoneid	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-force-override-listeners	':
		'',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-bandwidth	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-vswitch-id	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-additional-resource-tags	':
		'',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-ip-version	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-delete-protection	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-modification-protection	':
		'',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-resource-group-id	': '',
	'service.beta.kubernetes.io/alibaba-cloud-loadbalancer-name	': ''
};

const EKS = {
	'service.beta.kubernetes.io/load-balancer-source-ranges': '',
	'service.beta.kubernetes.io/aws-load-balancer-type	': '',
	'service.beta.kubernetes.io/aws-load-balancer-nlb-target-type': '',
	'service.beta.kubernetes.io/aws-load-balancer-name	': '',
	'service.beta.kubernetes.io/aws-load-balancer-internal': '',
	'service.beta.kubernetes.io/aws-load-balancer-scheme': '',
	'service.beta.kubernetes.io/aws-load-balancer-proxy-protocol': '',
	'service.beta.kubernetes.io/aws-load-balancer-ip-address-type': '',
	'service.beta.kubernetes.io/aws-load-balancer-access-log-enabled': '',
	'service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-name': '',
	'service.beta.kubernetes.io/aws-load-balancer-access-log-s3-bucket-prefix':
		'',
	'service.beta.kubernetes.io/aws-load-balancer-cross-zone-load-balancing-enabled':
		'',
	'service.beta.kubernetes.io/aws-load-balancer-ssl-cert': '',
	'service.beta.kubernetes.io/aws-load-balancer-ssl-ports': '',
	'service.beta.kubernetes.io/aws-load-balancer-ssl-negotiation-policy': '',
	'service.beta.kubernetes.io/aws-load-balancer-backend-protocol': '',
	'service.beta.kubernetes.io/aws-load-balancer-additional-resource-tags': '',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-healthy-threshold':
		'',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-unhealthy-threshold':
		'',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-timeout': '',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-interval': '',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-protocol': '',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-port	': '',
	'service.beta.kubernetes.io/aws-load-balancer-healthcheck-path': '',
	'service.beta.kubernetes.io/aws-load-balancer-eip-allocations': '',
	'service.beta.kubernetes.io/aws-load-balancer-private-ipv4-addresses	': '',
	'service.beta.kubernetes.io/aws-load-balancer-target-group-attributes': '',
	'service.beta.kubernetes.io/aws-load-balancer-subnets': '',
	'service.beta.kubernetes.io/aws-load-balancer-alpn-policy': '',
	'service.beta.kubernetes.io/aws-load-balancer-target-node-labels': ''
};

const OpenELB = {
	'lb.kubesphere.io/v1alpha1': '',
	'protocol.openelb.kubesphere.io/v1alpha1': '',
	'eip.openelb.kubesphere.io/v1alpha2': ''
};

export const CLUSTER_PROVIDERS_ANNOTATIONS = {
	'Aliyun ACK': aliyun,
	'Aure Kubernetes Service': microsoft,
	'Huawei Cloud CCE': Huawei,
	'Amazon EKS': EKS,
	'Google Kubernetes Engine': GoogleKubernetesEngine,
	'QingCloud Kubernetes Engine': QingCloud,
	'Tencent Kubernetes Engine': TencentKubernetesEngine,
	OpenELB
};
