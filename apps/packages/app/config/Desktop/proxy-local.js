module.exports = {
	'/server/myApps': {
		target: `http://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/search': {
		target: `http://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server': {
		target: `http://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/refresh': {
		target: `http://${'vault'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `http://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `http://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
