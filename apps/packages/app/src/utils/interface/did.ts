import { i18n } from 'src/boot/i18n';

const OlaresidError = {
	30215: {
		message: 'ORG_NOT_FOUND',
		txt: 'Organization not found, Check your Olares ID and try again'
	},
	30209: {
		message: 'ID_NOT_FOUND',
		txt: 'Incorrect Olares ID or password'
	}
} as const;

export const organizationsErrors = {
	'this org name is not available': 'This name is not available',
	'org name should be 1-63 characters, lowercase alphanumeric only':
		'Must be 2-24 characters, use lowercase letters and numbers only',
	'org name should be 2-24 characters, lowercase alphanumeric only':
		'Must be 2-24 characters, use lowercase letters and numbers only'
};

// this org name is not available

export const getOrganizationResponseErrorByCode = (
	code: number,
	message: string
) => {
	if (OlaresidError[code]) {
		return i18n.global.t(`organization_errors.${OlaresidError[code].txt}`);
	}
	return message;
};

export const getOrganizationResponseErrorMessage = (message: string) => {
	if (organizationsErrors[message]) {
		return i18n.global.t(`organization_errors.${organizationsErrors[message]}`);
	}
	return message;
};
