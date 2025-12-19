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

import { isObject } from 'lodash';
import yaml from 'js-yaml';

export const getValue = (value) => {
	if (isObject(value)) {
		try {
			return yaml.dump(value, { noRefs: true });
		} catch (err) {
			console.error(err);
			return JSON.stringify(value, null, 2);
		}
	}
	return String(value);
};

export const getValueObj = (value) => {
	if (!isObject(value)) {
		try {
			return yaml.load(value);
		} catch (err) {
			//
		}
	}
	return value;
};

export const getAllYAMLValue = (value) => {
	const objs = [];

	try {
		yaml.loadAll(
			value,
			(obj) => {
				objs.push(obj);
			},
			{ noRefs: true }
		);
	} catch (err) {
		//
	}

	return objs;
};
