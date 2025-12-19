const boot = ['i18n', 'marketUI', 'baseAxios', 'application/market'];
const css = ['market.scss'];

const getConfig = (ctx) => {
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Market/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};

	const proxyDefault = {
		'*': {
			target: process.env.PUBLIC_URL,
			changeOrigin: true
		}
	};

	const proxyConfig = process.env.PUBLIC_URL ? proxyDefault : undefined;

	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL,
				PUBLIC_URL: process.env.PUBLIC_URL,
				WS_URL: process.env.WS_URL,
				LOGIN_USERNAME: process.env.LOGIN_USERNAME,
				LOGIN_PASSWORD: process.env.LOGIN_PASSWORD
			},
			chainWebpack,
			distDir: 'dist/apps/market'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.market.html'
		},
		htmlVariables: {
			productName: process.env.PUBLIC_URL
				? 'Olares Application Marketplace'
				: 'Market'
		},
		animations: ['slideInDown', 'slideOutUp'],
		devServer: {
			proxy: proxyConfig,
			host: process.env.PUBLIC_URL ? 'localhost' : undefined,
			https: true
		}
	};
};

module.exports = {
	getConfig
};
