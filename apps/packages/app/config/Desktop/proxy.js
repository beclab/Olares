module.exports = {
	'/server/myApps': {
		target: `https://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/search': {
		target: `https://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server': {
		target: `https://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/refresh': {
		target: `https://${'vault'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/repos': {
		target: `https://${'files'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `https://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `https://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
