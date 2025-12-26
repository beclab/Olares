import { getAllYAMLValue, getValue } from '@apps/control-hub/src/utils/yaml';
import { isString } from 'lodash';

export const objectToYaml = (formTemplate: any) => {
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

export const yamlToObject = (data: any, hasMeta: boolean) => {
	const values = getAllYAMLValue(data);

	if (hasMeta && values.length === 1) {
		return values[0];
	}

	return values;
};
