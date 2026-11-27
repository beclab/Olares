export default {
	login_title: '登录 Olares',
	login_hint_password: '输入密码',
	mobile_veri: '在 LarePass 中确认',
	login_using_auth: '改用一次性密码验证',
	otp_title: '安全验证',
	otp_message: '输入 LarePass 中显示的一次性密码',
	login_errors: {
		'Authentication failed, incorrect password': '密码不正确',
		'Authentication failed, user not found': '未找到账户',
		'Authentication failed, lldap service is unavailable':
			'LLDAP 服务不可用，无法验证账户，请稍后重试',
		'Authentication failed, citus service is unavailable':
			'Citus 服务不可用，无法验证账户，请稍后重试',
		'Authentication failed, failed to query user from lldap service':
			'LLDAP 服务查询失败，无法找到账户，请稍后重试',
		'Authentication failed, disk space is full': '磁盘空间已满，无法登录',
		'Authentication failed. Check your credentials': '密码不正确',
		'too many failed login attempts, retry again later after 5 minutes':
			'登录尝试次数过多，5 分钟后重试'
	}
};
