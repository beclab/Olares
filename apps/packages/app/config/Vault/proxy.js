module.exports = {
	'/server': {
		target: `${process.env.PROTOCOL}vault.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/bfl/info/v1/olares-info': {
		target: `${process.env.PROTOCOL}vault.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
