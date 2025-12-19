module.exports = {
	'/api': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/videos': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/upload': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
