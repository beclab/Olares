const boot = ['i18n', 'smartEnginEntrance', 'baseAxios', 'application/share'];
const css = ['app.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Share/index');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};
	const proxy = require('./Share/proxy');
	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack,
			// distDir: 'dist/'
			distDir: 'dist/apps/share'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.share.html'
		},
		htmlVariables: {
			productName: 'Share'
		},
		devServer: {
			proxy: proxy,
			host: process.env.DEV_DOMAIN,
			https: true
		}
	};
};

module.exports = {
	getConfig
};
