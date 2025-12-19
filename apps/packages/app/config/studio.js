const boot = [
	'studioMonacoplugin',
	'studioMarkdown',
	'baseAxios',
	'studioUI',
	'studio-i18n',
	'application/studio'
];
const css = ['studio/app.scss'];

const proxyTarget = `${process.env.STUDIO_SUB_DOMAIN}.${process.env.ACCOUNT_DOMAIN}`;

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	return {
		boot,
		css,
		build: {
			env: {}
			// distDir: 'dist/apps/studio'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.studio.html',
			variables: 'studio/variables.scss'
		},
		htmlVariables: {
			productName: 'Studio'
		},
		// animations: 'all',
		devServer: {
			proxy: {
				'/api/command': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/apps': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/app-state': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/app-status': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/app-cfg': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/list-my-containers': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/files': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/api/command/list-app': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/upload': {
					// target: "http://127.0.0.1:3010/",
					target: `https://${proxyTarget}`,
					changeOrigin: true
				},
				'/socket.io': {
					target: 'ws://localhost:9000',
					ws: true
				},
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
				}
			}
		}
	};
};

module.exports = {
	getConfig
};
