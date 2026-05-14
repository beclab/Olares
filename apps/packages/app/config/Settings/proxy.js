module.exports = {
	'/api/logout': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/headscale': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/images': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/market-backend': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
