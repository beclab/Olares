module.exports = {
	'/api/env/appenv/remoteOptions': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/search': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/cookie/all': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/account/all': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/getUpgradeableVersion': {
		target: `${process.env.PROTOCOL}os-versions.api.api.jointerminus.cn`,
		changeOrigin: true
	},
	'/api/system/status': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/mdns/': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/abilities': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/myApps': {
		target: `${process.env.PROTOCOL}${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/myApps': {
		target: `${process.env.PROTOCOL}${'settings'}.${
			process.env.ACCOUNT_DOMAIN
		}`,
		changeOrigin: true
	},
	'/server/search': {
		target: `${process.env.PROTOCOL}${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server': {
		target: `${process.env.PROTOCOL}${
			process.env.APPLICATION === 'WISE'
				? process.env.WISE_SUB_DOMAIN
				: process.env.APPLICATION === 'SETTINGS'
				? 'settings'
				: process.env.APPLICATION === 'MARKET'
				? 'market'
				: process.env.APPLICATION === 'EDITOR' ||
				  process.env.APPLICATION === 'PREVIEW'
				? 'profile'
				: process.env.APPLICATION === 'LOGIN'
				? 'auth'
				: process.env.APPLICATION === 'DESKTOP'
				? 'desktop'
				: 'vault'
		}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/upload': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/olares-info': {
		target: `${process.env.PROTOCOL}${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/admin': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/apis/backup': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/headscale': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/images': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/myapps': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `${process.env.PROTOCOL}${
			process.env.APPLICATION === 'WISE'
				? process.env.WISE_SUB_DOMAIN
				: process.env.APPLICATION === 'SETTINGS' ||
				  process.env.APPLICATION === 'MARKET'
				? 'settings'
				: process.env.APPLICATION === 'EDITOR' ||
				  process.env.APPLICATION === 'PREVIEW'
				? 'profile'
				: process.env.APPLICATION === 'LOGIN'
				? 'auth'
				: process.env.APPLICATION === 'DESKTOP'
				? 'desktop'
				: 'files'
		}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/second': {
		target: `${process.env.PROTOCOL}auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/files': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/seahub': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/seafhttp': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/settingsApi': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/settingsApi': ''
		}
	},
	'/videos': {
		target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/knowledge/api/cookie': {
		target: `${process.env.PROTOCOL}settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/knowledge': ''
		}
	},
	'/knowledge': {
		target: `${process.env.PROTOCOL}${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/artifact-files': {
		target: `${process.env.PROTOCOL}${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/download': {
		target: `${process.env.PROTOCOL}${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/desktop': {
		target: `${process.env.PROTOCOL}desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/desktop': ''
		}
	},
	'/app-store': {
		target: process.env.PUBLIC_URL
			? process.env.PUBLIC_URL
			: `${process.env.PROTOCOL}market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
