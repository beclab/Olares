module.exports = {
	'/api': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/videos': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/upload': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/seafhttp': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
