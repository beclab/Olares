const boot = ['i18n', 'desktopUI', 'baseAxios', 'application/desktop'];
const css = ['app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Desktop/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};

	const proxy = require('./Desktop/proxy');
	const proxyLocal = require('./Desktop/proxy-local');

	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL
			},
			chainWebpack,
			distDir: 'dist/apps/desktop'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.desktop.html'
		},
		htmlVariables: {
			productName: 'Olares'
		},
		animations: ['fadeIn', 'fadeOut'],
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: true
		},
		pwa: {
			workboxPluginMode: 'InjectManifest', // 'GenerateSW' or 'InjectManifest'
			workboxOptions: {}, // only for GenerateSW

			manifest: {
				name: 'Olares',
				short_name: 'Desktop',
				description: 'Olares OS Launcher',
				display: 'standalone',
				orientation: 'portrait',
				theme_color: 'transparent',
				icons: [
					{
						src: 'desktop/icons/icon-128x128.png',
						sizes: '128x128',
						type: 'image/png'
					},
					{
						src: 'desktop/icons/icon-192x192.png',
						sizes: '192x192',
						type: 'image/png'
					},
					{
						src: 'desktop/icons/icon-256x256.png',
						sizes: '256x256',
						type: 'image/png'
					},
					{
						src: 'desktop/icons/icon-384x384.png',
						sizes: '384x384',
						type: 'image/png'
					},
					{
						src: 'desktop/icons/icon-512x512.png',
						sizes: '512x512',
						type: 'image/png'
					}
				]
			}
		}
	};
};

module.exports = {
	getConfig
};
