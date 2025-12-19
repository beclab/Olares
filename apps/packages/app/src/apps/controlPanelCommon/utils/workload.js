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

import { get, isEmpty, set, merge, isUndefined } from 'lodash';

import { to, cpuFormat, memoryFormat, resourceLimitKey } from './index.js';
import { getVolumeType } from './volume.js';
import { ObjectMapper } from './object.mapper';
import { api } from '@apps/control-panel-common/src/boot/axios';

export const getHpaFormattedData = (formData = {}) => {
	const cpuCurrentUtilization = get(
		formData,
		'metadata.annotations.cpuCurrentUtilization',
		0
	);
	const cpuTargetUtilization = get(
		formData,
		'metadata.annotations.cpuTargetUtilization',
		0
	);
	const memoryCurrentValue = get(
		formData,
		'metadata.annotations.memoryCurrentValue',
		0
	);
	const memoryTargetValue = get(
		formData,
		'metadata.annotations.memoryTargetValue',
		''
	);
	const cpuMetric = cpuTargetUtilization
		? [
				{
					type: 'Resource',
					resource: {
						name: 'cpu',
						target: {
							type: 'Utilization',
							averageUtilization: Number(cpuTargetUtilization)
						}
					}
				}
		  ]
		: [];
	const memoryMetric = memoryTargetValue
		? [
				{
					type: 'Resource',
					resource: {
						name: 'memory',
						target: {
							type: 'AverageValue',
							averageValue: memoryTargetValue
						}
					}
				}
		  ]
		: [];

	const data = merge(formData, {
		apiVersion: 'autoscaling/v2beta2',
		kind: 'HorizontalPodAutoscaler',
		metadata: {
			annotations: {
				cpuCurrentUtilization: String(cpuCurrentUtilization),
				cpuTargetUtilization: String(cpuTargetUtilization),
				memoryCurrentValue: String(memoryCurrentValue),
				memoryTargetValue: String(memoryTargetValue)
			}
		}
	});

	set(data, 'spec.metrics', [...memoryMetric, ...cpuMetric]);

	return data;
};

export const getWorkloadVolumes = async (detail, setDetail = false) => {
	const prefix = detail.cluster
		? `api/v1/klusters/${detail.cluster}/namespaces/${detail.namespace}/persistentvolumeclaims`
		: `api/v1/namespaces/${detail.namespace}/persistentvolumeclaims`;
	let specVolumes = [];
	if (!isEmpty(detail.volumes)) {
		const promises = [];
		specVolumes = detail.volumes.map((volume) => {
			let volumeName = '';
			if (volume.persistentVolumeClaim) {
				volumeName = volume.persistentVolumeClaim.claimName;
				promises.push(to(api.get(`${prefix}/${volumeName}`)));
			} else {
				volumeName = volume.HostPath ? volume.HostPath.path : volume.name;
			}

			return {
				...volume,
				storage: '-',
				name: volumeName,
				mountName: volume.name,
				type: getVolumeType(volume)
			};
		});

		const volumesResult = await Promise.all(promises);
		const volumes = volumesResult.map(ObjectMapper.volumes);

		specVolumes.forEach((volume) => {
			if (volume.type === 'Volume') {
				const volumeDetail =
					volumes.find((_volume) => _volume.name === volume.name) || {};
				volume.storage = volumeDetail.capacity;
				volume.capacity = volumeDetail.capacity;
				volume.storageClassName = volumeDetail.storageClassName;
				volume.accessMode = get(volumeDetail, 'accessMode');
			}

			// get volume mounts
			if (!isEmpty(detail.containers)) {
				const mounts = [];

				detail.containers.forEach((container) => {
					if (!isEmpty(container.volumeMounts)) {
						container.volumeMounts.forEach((mount) => {
							if (mount.name === volume.mountName) {
								mounts.push({
									...mount,
									container: container.name,
									accessMode: mount.readOnly ? 'read-only' : 'read-write'
								});
							}
						});
					}
				});

				volume.mounts = mounts;
			}
		});
	}

	if (setDetail) {
		detail.volumes = specVolumes;
	}

	return specVolumes;
};

export const getWorkloadUpdateTime = (item) => {
	const annotations = get(item, 'metadata.annotations', {});
	const status = get(item, 'status', {});

	if (annotations.updateTime) return annotations.updateTime;

	const conditions = status.conditions || [];

	if (isEmpty(conditions)) return get(item, 'metadata.creationTimestamp');

	let lastTime = new Date(get(conditions, '[0].lastUpdateTime', 0)).valueOf();
	conditions.forEach(({ lastUpdateTime }) => {
		const value = new Date(lastUpdateTime).valueOf();
		value > lastTime && (lastTime = value);
	});

	return lastTime;
};

export const getJobUpdateTime = (item) => {
	const status = get(item, 'status', {});

	const conditions = status.conditions || [];

	if (isEmpty(conditions)) return get(item, 'metadata.creationTimestamp');

	let lastTime = new Date(
		get(conditions, '[0].lastTransitionTime', 0)
	).valueOf();
	conditions.forEach(({ lastTransitionTime }) => {
		const value = new Date(lastTransitionTime).valueOf();
		value > lastTime && (lastTime = value);
	});

	return lastTime;
};

export const getCurrentRevision = (
	workloadDetail,
	revisions,
	module = 'deployments'
) => {
	let revision = 0;

	switch (module) {
		default:
		case 'deployments':
			revision = get(
				workloadDetail,
				'annotations["deployment.kubernetes.io/revision"]'
			);
			break;
		case 'statefulsets':
		case 'daemonsets': {
			let maxRevision = get(revisions[0], 'revision', 0);
			for (let i = 1; i < revisions.length; i++) {
				if (revisions[i].revision > maxRevision) {
					maxRevision = revisions[i].revision;
				}
			}

			const cur = revisions.find(
				(item) => item.name === get(workloadDetail, 'status.currentRevision')
			);
			revision = cur ? cur.revision : maxRevision;
			break;
		}
	}

	return parseInt(revision, 10);
};

export const getWorkloadReplicaCount = (record, module) => {
	const result = { ready: 0, total: 0 };

	switch (module) {
		default:
		case 'jobs':
		case 'cronjobs':
		case 'deployments':
		case 'statefulsets':
			result.ready = get(record, 'readyPodNums', 0);
			result.total = get(record, 'podNums', 0);
			break;
		case 'daemonsets':
			result.ready = get(record, 'status.numberReady', 0);
			result.total = get(record, 'status.desiredNumberScheduled', 0);
			break;
	}

	return result;
};

export const getLeftQuota = (wsQuota, nsQuota) => {
	function getLeft(quota, key) {
		let left;
		const total = get(quota, `hard["${key}"]`);
		const used = get(quota, `used["${key}"]`, '0');
		if (total) {
			if (key.endsWith('cpu')) {
				left = cpuFormat(total) - cpuFormat(used);
			} else if (key.endsWith('memory')) {
				left = memoryFormat(total) - memoryFormat(used);
			}
		}

		return left;
	}

	return {
		workspace: resourceLimitKey.reduce(
			(prev, key) => ({ ...prev, [key]: getLeft(wsQuota, key) }),
			{}
		),
		namespace: resourceLimitKey.reduce(
			(prev, key) => ({ ...prev, [key]: getLeft(nsQuota, key) }),
			{}
		)
	};
};

export const getUsedQuota = (data) => {
	function getQuota(quota, key) {
		let result;
		const used = get(quota, `used["${key}"]`);
		if (used !== '0') {
			if (key.endsWith('cpu')) {
				result = cpuFormat(used);
			} else if (key.endsWith('memory')) {
				result = memoryFormat(used);
			}
		}

		return result;
	}

	const quotas = resourceLimitKey.reduce(
		(prev, key) => ({ ...prev, [key]: getQuota(data, key) }),
		{}
	);
	return { ...quotas };
};

export const getContainersResources = (
	containers,
	initContainers,
	replicas = 1
) => {
	const keys = [
		'limits.cpu',
		'limits.memory',
		'requests.cpu',
		'requests.memory'
	];

	const resources = {};
	keys.forEach((key) => {
		let initResource = 0;
		let resource = 0;

		let formatter = (v) => v;
		if (key.endsWith('cpu')) {
			formatter = cpuFormat;
		} else if (key.endsWith('memory')) {
			formatter = memoryFormat;
		}

		// get the max init container request
		initContainers.forEach((item) => {
			const value = formatter(get(item, `resources.${key}`));
			if (isUndefined(value) || value === null) {
				initResource = Infinity;
			} else if (initResource !== Infinity) {
				initResource = Math.max(value, initResource);
			}
		});

		// sum all the containers' request
		containers.forEach((item) => {
			const value = formatter(get(item, `resources.${key}`));
			if (isUndefined(value) || value === null) {
				resource = Infinity;
			} else if (resource !== Infinity) {
				resource += value;
			}
		});

		resources[key] = replicas * Math.max(initResource, resource);
	});

	return resources;
};

export const compareQuotaAndResources = (leftQuota, resources) => {
	const result = {};
	const keys = Object.keys(resources);
	keys.forEach((key) => {
		const cost = resources[key];
		let unit = '';
		const workspaceQuota = get(leftQuota, `workspace["${key}"]`, Infinity);
		const namespaceQuota = get(leftQuota, `namespace["${key}"]`, Infinity);

		if (key.endsWith('cpu')) {
			unit = 'Core';
		} else if (key.endsWith('memory')) {
			unit = 'Mi';
		}

		result[key] = {
			cost: cost === Infinity ? undefined : `${cost}${unit}`,
			namespaceQuota:
				namespaceQuota === Infinity ? undefined : `${namespaceQuota}${unit}`,
			workspaceQuota:
				workspaceQuota === Infinity ? undefined : `${workspaceQuota}${unit}`,
			overcost:
				cost === Infinity
					? 'unset'
					: cost > workspaceQuota || cost > namespaceQuota
		};
	});

	return result;
};
