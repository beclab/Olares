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

import { get, isEmpty } from 'lodash';
import { SERVICE_TYPES } from './constants';

export const getServiceType = (item) => {
	const specType = get(item, 'spec.type');
	const clusterIP = get(item, 'spec.clusterIP');
	const selector = get(item, 'spec.selector', {});

	let type = SERVICE_TYPES.VirtualIP;
	if (specType === 'ClusterIP') {
		if (clusterIP === 'None' || clusterIP === '') {
			if (!isEmpty(selector)) {
				type = SERVICE_TYPES.Headless;
			} else {
				type = SERVICE_TYPES.Unknown;
			}
		}
	} else if (specType === 'ExternalName') {
		type = SERVICE_TYPES.ExternalName;
	}

	return type;
};

export const getServicePort = (item) => {
	let { protocol } = item;
	if (item.name && item.name.indexOf('-') !== -1) {
		protocol = item.name.split('-')[0].toUpperCase();
	}

	return `${item.port}${
		item.targetPort ? `:${item.targetPort}` : ''
	}/${protocol}`;
};

export const getMetricData = (values, defaultValue) => {
	if (values[values.length - 1]) {
		return Number(values[values.length - 1][1]) || defaultValue;
	}

	return defaultValue;
};

export const getSuccessCount = (total = [], error = []) => {
	const totalNum = Number(total[1]) || 0;
	const errorNum = Number(error[1]) || 0;

	return [total[0], totalNum > 0 ? totalNum - errorNum : NaN];
};

export const getSuccessRate = (total = [], error = []) => {
	const totalNum = Number(total[1]) || 0;
	const errorNum = Number(error[1]) || 0;

	return [
		total[0],
		totalNum > 0 ? ((totalNum - errorNum) / totalNum).toFixed(4) : NaN
	];
};
