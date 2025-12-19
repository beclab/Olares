module.exports = {
	'/bfl': {
		target: `https://auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `https://auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
