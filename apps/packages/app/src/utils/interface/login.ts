export const loginErrors = [
	'Authentication failed, incorrect password',
	'Authentication failed, user not found',
	'Authentication failed, lldap service is unavailable',
	'Authentication failed, citus service is unavailable',
	'Authentication failed, failed to query user from lldap service',
	'Authentication failed, disk space is full',
	'Authentication failed. Check your credentials',
	'too many failed login attempts, retry again later after 5 minutes'
];
export const getLoginResponseErrorMessage = (message: string) => {
	const errorIndex = loginErrors.findIndex((e) => message.startsWith(e));
	if (errorIndex >= 0) {
		return `login_errors.${loginErrors[errorIndex]}`;
	}
	return message;
};
