module.exports = {
	'/api': {
		target: `https://share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/upload': {
		target: `https://share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/bfl': {
		target: `https://share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
