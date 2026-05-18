module.exports = {
	'/bfl': {
		target: `${process.env.PROTOCOL}auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `${process.env.PROTOCOL}auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
