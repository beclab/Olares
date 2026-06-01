module.exports = {
	'/api': {
		target: `${process.env.PROTOCOL}share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/upload': {
		target: `${process.env.PROTOCOL}share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/bfl': {
		target: `${process.env.PROTOCOL}share.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
