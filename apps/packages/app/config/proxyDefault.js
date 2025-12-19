module.exports = {
	'/api/env/appenv/remoteOptions': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/search': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/cookie/all': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/account/all': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/getUpgradeableVersion': {
		target: `https://os-versions.api.api.jointerminus.cn`,
		changeOrigin: true
	},
	'/api/system/status': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/mdns/': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/abilities': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/myApps': {
		target: `https://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server/search': {
		target: `https://${'desktop'}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/server': {
		target: `https://${
			process.env.APPLICATION === 'WISE'
				? process.env.WISE_SUB_DOMAIN
				: process.env.APPLICATION === 'SETTINGS'
				? 'settings'
				: process.env.APPLICATION === 'MARKET'
				? 'market'
				: process.env.APPLICATION === 'EDITOR' ||
				  process.env.APPLICATION === 'PREVIEW'
				? 'profile'
				: process.env.APPLICATION === 'DESKTOP'
				? 'desktop'
				: 'vault'
		}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/upload': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/olares-info': {
		target: `https://${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/admin': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/apis/backup': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/kapis': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/headscale': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/images': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api/myapps': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/api': {
		target: `https://${
			process.env.APPLICATION === 'WISE'
				? process.env.WISE_SUB_DOMAIN
				: process.env.APPLICATION === 'SETTINGS'
				? 'settings'
				: process.env.APPLICATION === 'EDITOR' ||
				  process.env.APPLICATION === 'PREVIEW'
				? 'profile'
				: process.env.APPLICATION === 'DESKTOP'
				? 'desktop'
				: 'files'
		}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/second': {
		target: `https://auth.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/files': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/seahub': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/seafhttp': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/settingsApi': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/settingsApi': ''
		}
	},
	'/videos': {
		target: `https://files.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		secure: false
	},
	'/knowledge/api/cookie': {
		target: `https://settings.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/knowledge': ''
		}
	},
	'/knowledge': {
		target: `https://${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/artifact-files': {
		target: `https://${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/download': {
		target: `https://${process.env.WISE_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	},
	'/desktop': {
		target: `https://desktop.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true,
		pathRewrite: {
			'^/desktop': ''
		}
	},
	'/app-store': {
		target: process.env.PUBLIC_URL
			? process.env.PUBLIC_URL
			: `https://market.${process.env.ACCOUNT_DOMAIN}`,
		changeOrigin: true
	}
};
