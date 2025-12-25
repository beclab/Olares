/* eslint-disable no-dupe-else-if */
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

import { get, isUndefined, isEmpty, sortBy, includes } from 'lodash';

const POD_LIFE_CONDITIONS = [
	'Initialized',
	'Ready',
	'ContainersReady',
	'PodScheduled'
];

export const getDeployStatus = ({
	hasS2i,
	spec = {},
	status = {},
	annotations = {}
}) => {
	if (hasS2i) {
		const s2iStatus = get(
			annotations,
			"['devops.kubesphere.io/inithasbeencomplted']",
			'Running'
		);
		if (s2iStatus !== 'Successful') {
			return `S2I_${s2iStatus}`;
		}
	}

	if (
		spec.replicas === 0 &&
		!isUndefined(status.readyReplicas) &&
		status.readyReplicas !== 0
	) {
		return 'Updating';
	}

	if (spec.replicas) {
		if (status.readyReplicas === spec.replicas) {
			return 'Running';
		}
		return 'Updating';
	}

	return 'Stopped';
};

export const getStatefulSetStatus = ({ spec = {}, status = {} }) => {
	if (
		spec.replicas === 0 &&
		!isUndefined(status.readyReplicas) &&
		status.readyReplicas !== 0
	) {
		return 'Updating';
	}

	if (spec.replicas) {
		if (status.readyReplicas === spec.replicas) {
			return 'Running';
		}
		return 'Updating';
	}

	return 'Stopped';
};

export const getDaemonSetStatus = ({ status }) => {
	if (status.desiredNumberScheduled === 0) {
		return 'Stopped';
	}
	if (status.numberAvailable === status.desiredNumberScheduled) {
		return 'Running';
	}
	return 'Updating';
};

export const getJobStatus = ({ spec, status }) => {
	if (isEmpty(status)) return 'Failed';

	let _status = 'RUNNING';

	if (!isEmpty(status.conditions)) {
		status.conditions.some((item) => {
			if (item.type === 'Failed' && item.status === 'True') {
				_status = 'FAILED';
				return true;
			}

			if (item.type === 'Complete' && item.status === 'True') {
				_status = 'COMPLETED';
				return true;
			}

			return false;
		});
	}

	if (spec.Completions <= status.Succeed) _status = 'Failed';

	return _status;
};

export const getCronJobStatus = ({ spec }) => {
	if (spec && spec.suspend) {
		return 'Paused';
	}
	return 'Running';
};

export const getVolumeStatus = ({ status = {} }) => {
	const conditions = status.conditions || [];
	if (!isEmpty(conditions)) {
		const detailedStatus =
			conditions.find(({ type }) =>
				includes(['FileSystemResizePending', 'Resizing'], type)
			) || {};

		return detailedStatus.type;
	}
};

export const getWorkloadStatus = (record, module) => {
	const funcMap = {
		deployments: getDeployStatus,
		statefulsets: getStatefulSetStatus,
		daemonsets: getDaemonSetStatus,
		jobs: getJobStatus,
		cronjobs: getCronJobStatus
	};

	let status = '';
	let reason = '';

	if (funcMap[module]) {
		status = funcMap[module](record);
	}

	if (status === 'Updating') {
		const conditions = get(record, 'status.conditions', [])
			.filter((cd) =>
				cd.type === 'ReplicaFailure'
					? cd.status === 'True'
					: cd.status === 'False'
			)
			.reverse();
		reason = get(conditions, '[0].reason');
	}

	return { status, reason };
};

export const getContainerStatus = ({ state = {}, ready } = {}) => {
	let status = 'waiting';
	const keys = Object.keys(state);

	if (keys.length === 0) {
		status = 'waiting';
	} else if (keys.length === 1) {
		status = keys[0];
	} else {
		status = keys[0];
		keys.forEach((key) => {
			if (state[status].startedAt < state[key].startedAt) {
				status = key;
			}
		});
	}

	let reason = get(state, `${status}.reason`, '');

	if (!ready && status === 'running') {
		status = 'waiting';
		reason = 'ContainerNotReady';
	}

	return { reason, status };
};

export const getPodStatusAndRestartCount = (pod) => {
	const {
		phase = '',
		reason = '',
		containerStatuses,
		initContainerStatuses = [],
		conditions = []
	} = pod.status || {};
	let status = phase;
	let restarts = 0;
	let initializing = false;
	let readyCondition;

	if (!isEmpty(reason)) {
		status = reason;
	}

	if (!isEmpty(initContainerStatuses)) {
		initContainerStatuses.forEach((container, index) => {
			restarts += Number(container.restartCount);
			const waiting = get(container, 'state.waiting');
			const terminated = get(container, 'state.terminated');
			if (terminated && terminated.exitCode === 0) {
				return;
			}

			if (terminated) {
				if (isEmpty(terminated.reason)) {
					status = `Init:${terminated.reason}`;
				} else if (terminated.signal !== 0) {
					status = `Init:Signal:${terminated.signal}`;
				} else {
					status = `Init:ExitCode:${terminated.exitCode}`;
				}
				initializing = true;
			} else if (
				waiting &&
				terminated &&
				!isEmpty(terminated.reason) &&
				waiting.reason !== 'PodInitializing'
			) {
				status = `Init:${terminated.reason}`;
				initializing = true;
			} else {
				const len = get(pod, 'spec.initContainers.length');
				status = `Init:${index}/${len}`;
				initializing = true;
			}
		});
	}

	if (!initializing && !isEmpty(containerStatuses)) {
		let hasRunning = false;
		restarts = 0;
		containerStatuses.forEach((container) => {
			restarts += Number(container.restartCount);

			const waiting = get(container, 'state.waiting');
			const terminated = get(container, 'state.terminated');

			if (waiting && waiting.reason) {
				status = waiting.reason;
			} else if (terminated && terminated.reason) {
				status = terminated.reason;
			} else if (terminated && !terminated.reason) {
				status =
					terminated.singal !== 0
						? `Signal:${terminated.singal}`
						: `ExitCode:${terminated.exitCode}`;
			} else if (
				container.ready &&
				container.state &&
				container.state.running
			) {
				hasRunning = true;
			}
		});

		if (status === 'Completed' && hasRunning) {
			status = 'Running';
		}

		const _conditions = conditions.filter(
			(item) => POD_LIFE_CONDITIONS.indexOf(item.type) !== -1
		);

		_conditions.forEach((item) => {
			if (status === 'Running') {
				if (
					item.type === 'Unschedulable'
						? item.status === 'True'
						: item.status === 'False'
				) {
					status = 'Pending';
				}
			}

			if (item.type === 'Ready') {
				readyCondition = item;
			}
		});
	}

	if (get(pod, 'metadata.deletionTimestamp')) {
		status = reason === 'NodeLost' ? 'UnKnown' : 'Terminating';
	}

	let type = 'waiting';
	switch (status) {
		case 'Running':
			type = 'running';
			break;
		case 'Failed':
		case 'Error':
			type = 'error';
			break;
		case 'Completed':
			type = 'completed';
			break;
		case 'Succeeded':
			if (readyCondition.reason === 'PodCompleted') {
				type = 'completed';
			} else {
				status = readyCondition.reason;
			}
			break;
		case 'Pending':
			{
				const cds = sortBy(
					conditions.filter((item) =>
						item.type === 'Unschedulable'
							? item.status === 'True'
							: item.status === 'False'
					),
					'lastTransitionTime'
				);
				status = get(cds, '[0].reason', status);
			}
			break;
		default:
			if (status.startsWith('OutOf')) {
				type = 'error';
			}
			// Init:*, Terminating, Signal:*, ExitCode:*, ContainerCreating, Others
			break;
	}

	return { status, type, restarts };
};

export const getComponentStatus = ({ totalBackends, healthyBackends }) => {
	if (healthyBackends !== totalBackends) {
		return 'Warning';
	}

	if (totalBackends === 0) {
		return 'Stopped';
	}

	return 'Healthy';
};
