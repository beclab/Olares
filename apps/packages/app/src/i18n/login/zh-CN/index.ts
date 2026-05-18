export default {
	login_title: '输入密码登录',
	login_hint_password: '密码',
	mobile_veri: '在 LarePass 应用程序上确认',
	login_using_auth: '使用 LarePass 生成的一次性密码验证',
	otp_title: '安全验证',
	otp_message: '输入 LarePass 的一次性密码',
	login_errors: {
		'Authentication failed, incorrect password': '登录失败，密码错误。',
		'Authentication failed, user not found': '登录失败，用户未找到。',
		'Authentication failed, lldap service is unavailable':
			'登录失败，lldap服务异常。',
		'Authentication failed, citus service is unavailable':
			'登录失败，citus服务异常。',
		'Authentication failed, failed to query user from lldap service':
			'登录失败，lldap未查询到该用户。',
		'Authentication failed, disk space is full': '登录失败，磁盘空间已满。',
		'Authentication failed. Check your credentials': '登录失败，密码错误。',
		'too many failed login attempts, retry again later after 5 minutes':
			'登录尝试次数过多，请5分钟后重试。'
	}
};
