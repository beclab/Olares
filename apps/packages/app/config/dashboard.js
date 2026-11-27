const boot = [
	'dashboard-i18n',
	'baseAxios',
	'dashboardUI',
	'application/dashboard',
	'adaptive-body-class'
];
const css = ['dashboard/app.scss', 'quasar-layout.scss'];

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
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/kapis': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/api/mdns/': {
					target: `${process.env.PROTOCOL}settings.${proxyTarget}`,
					changeOrigin: true
				},
				'/api/gpu/list': {
					target: `${process.env.PROTOCOL}settings.${proxyTarget}`,
					changeOrigin: true
				},
				'/api': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/bfl': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/server': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/analytics_service': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				},
				'/capi': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/user-service': {
					target: `${process.env.PROTOCOL}dashboard.${proxyTarget}`,
					changeOrigin: true
				}
			}
		}
	};
};

module.exports = {
	getConfig
};
