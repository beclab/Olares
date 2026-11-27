module.exports = {
	'/app-store/': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/env': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/backend': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/users': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/compute-resources': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/apps/**/compute-resources/bindings': {
		target: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
