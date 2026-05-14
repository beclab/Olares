module.exports = {
	'/api': {
		target: `${process.env.PROTOCOL}auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/bfl': {
		target: `${process.env.PROTOCOL}auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
