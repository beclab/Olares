const boot = [
	'control-hub-i18n',
	'baseAxios',
	'controlHubUI',
	'control-hub-permission',
	'application/controlHub'
];
const css = ['controlHub/app.scss'];

const proxyTarget = `control-hub.${process.env.ACCOUNT_DOMAIN}`;

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	return {
		boot,
		css,
		build: {
			env: {},
			distDir: 'dist/apps/control-hub'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.control_hub.html',
			variables: 'controlHub/variables.scss'
		},
		htmlVariables: {
			productName: 'Control-Hub'
		},
		animations: ['fadeInRight', 'fadeIn', 'fadeInRight', 'fadeInRight'],
		devServer: {
			proxy: {
				'/kapis/terminal': {
					target: `wss://${proxyTarget}`,
					changeOrigin: true,
					ws: true,
					http: false
				},
				'/apis/apps/v1/watch': {
					target: `wss://${proxyTarget}`,
					changeOrigin: true,
					ws: true,
					http: false
				},
				'/api/v1/watch': {
					target: `wss://${proxyTarget}`,
					changeOrigin: true,
					ws: true,
					http: false
				},
				'/kapis': {
					target: `https://${proxyTarget}`,
					changeOrigin: true,
					secure: false,
					ws: false
				},
				'/api': {
					target: `https://${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/bfl': {
					target: `https://${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/capi': {
					target: `https://${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/middleware': {
					target: `https://${proxyTarget}`,
					changeOrigin: true,
					secure: false
				},
				'/user-service': {
					target: `https://${proxyTarget}`,
					changeOrigin: true
				}
			}
		}
	};
};

module.exports = {
	getConfig
};
