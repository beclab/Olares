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

import { isObject, isString } from 'lodash';
import yaml from 'js-yaml';

export const getValue = (value) => {
	if (isObject(value)) {
		try {
			return yaml.dump(JSON.parse(JSON.stringify(value)), { noRefs: true });
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

export const objectToYaml = (formTemplate) => {
	if (formTemplate.metadata) {
		return getValue(formTemplate);
	}

	if (isString(formTemplate)) {
		return formTemplate;
	}

	return Object.values(formTemplate)
		.map((value) => getValue(value || {}))
		.join('---\n');
};

export const yamlToObject = (data, hasMeta) => {
	const values = getAllYAMLValue(data);

	if (hasMeta && values.length === 1) {
		return values[0];
	}

	return values;
};
