module.exports = {
	'/server/myApps': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/search': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/market-backend': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/refresh': {
		target: `${process.env.PROTOCOL}vault.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/repos': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
