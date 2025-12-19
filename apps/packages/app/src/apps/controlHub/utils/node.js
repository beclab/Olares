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

import { isEmpty } from 'lodash';

export const getNodeRoles = (labels) => {
	let roles = [];

	if (!isEmpty(labels)) {
		roles = Object.keys(labels)
			.filter((key) => key.startsWith('node-role.kubernetes.io/'))
			.map((key) => key.replace('node-role.kubernetes.io/', ''));
	}

	return roles;
};

export const NODE_CONDITION_ICONS = {
	Ready: 'templet',
	OutOfDisk: 'storage',
	PIDPressure: 'pie-chart',
	MemoryPressure: 'memory',
	DiskPressure: 'storage',
	NetworkUnavailable: 'earth',
	ConfigOK: 'record',
	KubeletReady: 'cluster'
};

export const getConditionsStatus = (record) => {
	if (record.status === 'Unknown') {
		return 'Warning';
	}

	switch (record.type) {
		case 'OutOfDisk':
			if (record.status === 'True') return 'Warning';
			break;
		case 'MemoryPressure':
			if (record.status === 'True') return 'Warning';
			break;
		case 'DiskPressure':
			if (record.status === 'True') return 'Warning';
			break;
		case 'PIDPressure':
			if (record.status === 'True') return 'Warning';
			break;
		case 'NetworkUnavailable':
			if (record.status === 'True') return 'Warning';
			break;
		case 'ConfigOK':
			if (record.status === 'False') return 'Warning';
			break;
		case 'KubeletReady':
			if (record.status === 'False') return 'Warning';
			break;
		case 'Ready':
			if (record.status !== 'True') return 'Warning';
			break;
		default:
			break;
	}

	return 'Running';
};

export const getNodeStatus = ({ status = {}, spec = {}, importStatus }) => {
	if (importStatus && importStatus !== 'success') {
		return importStatus;
	}

	const conditions = status.conditions || [];
	let health = true;

	if (spec.unschedulable) {
		return 'Unschedulable';
	}

	conditions.forEach((item) => {
		health = getConditionsStatus(item) === 'Running';
	});

	return health ? 'Running' : 'Warning';
};
