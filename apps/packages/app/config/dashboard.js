const boot = [
	'dashboard-i18n',
	'baseAxios',
	'dashboardUI',
	'application/dashboard'
];
const css = ['dashboard/app.scss'];

const proxyTarget = process.env.ACCOUNT_DOMAIN;

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	return {
		boot,
		css,
		build: {
			env: {},
			distDir: 'dist/apps/dashboard'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.dashboard.html',
			variables: 'dashboard/variables.scss'
		},
		htmlVariables: {
			productName: 'Dashboard'
		},
		animations: ['fadeInRight', 'fadeOutRight'],
		devServer: {
			proxy: {
				'/hami': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/kapis': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/api/mdns/': {
					target: `https://settings.${proxyTarget}`,
					changeOrigin: true
				},
				'/api/gpu/list': {
					target: `https://settings.${proxyTarget}`,
					changeOrigin: true
				},
				'/api': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/bfl': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/server': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/analytics_service': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/capi': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/user-service': {
					target: `https://dashboard.${proxyTarget}`,
					changeOrigin: true
				}
			}
		}
	};
};

module.exports = {
	getConfig
};
