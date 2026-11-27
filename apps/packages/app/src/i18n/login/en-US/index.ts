export default {
	login_title: 'Log in to Olares',
	login_hint_password: 'Enter password',
	mobile_veri: 'Confirm in LarePass',
	login_using_auth: 'Use one-time password instead',
	otp_title: 'Security verification',
	otp_message: 'Enter the one-time password shown in LarePass',
	login_errors: {
		'Authentication failed, incorrect password': 'Incorrect password',
		'Authentication failed, user not found': 'Account not found',
		'Authentication failed, lldap service is unavailable':
			'Unable to verify your account because the LLDAP service is unavailable. Try again later.',
		'Authentication failed, citus service is unavailable':
			'Unable to verify your account because the Citus service is unavailable. Try again later.',
		'Authentication failed, failed to query user from lldap service':
			'Unable to find your account because the LLDAP service query failed. Try again later.',
		'Authentication failed, disk space is full':
			'Unable to log in because disk space is full',
		'Authentication failed. Check your credentials': 'Incorrect password',
		'too many failed login attempts, retry again later after 5 minutes':
			'Too many login attempts. Try again in 5 minutes.'
	}
};
