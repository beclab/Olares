const { OLARES_VERSION } = require('./publicMarket');

const boot = ['i18n', 'marketUI', 'baseAxios', 'application/market'];
const css = ['market/app.scss', 'quasar-layout.scss'];

const getConfig = (ctx) => {
	// A public build is answered for one Olares version, which the config above
	// owns. The environment still wins, so a release can be tried against
	// another version without editing anything.
	const olaresVersion = process.env.OLARES_VERSION || OLARES_VERSION;
	if (process.env.PUBLIC_URL && !olaresVersion) {
		throw new Error('OLARES_VERSION is required for a public Market build');
	}
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}
	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpackMobile = require('./Market/config');
		chainWebpackMobile.build.chainWebpack(ctx, chain, { isClient });
	};

	const proxyDefault = {
		'**': {
			target: process.env.PUBLIC_URL,
			changeOrigin: true
		}
	};

	const proxyFile = require('./Market/proxy');

	const proxyConfig = process.env.PUBLIC_URL ? proxyDefault : proxyFile;

	return {
		boot,
		css,
		build: {
			env: {
				URL: process.env.URL,
				PUBLIC_URL: process.env.PUBLIC_URL,
				OLARES_VERSION: olaresVersion,
				WS_URL: process.env.WS_URL,
				LOGIN_USERNAME: process.env.LOGIN_USERNAME,
				LOGIN_PASSWORD: process.env.LOGIN_PASSWORD
			},
			chainWebpack,
			distDir: 'dist/apps/market'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.market.html',
			variables: 'market/variables.scss'
		},
		htmlVariables: {
			productName: process.env.PUBLIC_URL
				? 'Olares Application Marketplace' // i18n-ignore -- build-time HTML product name, not runtime i18n
				: 'Market'
		},
		animations: ['slideInDown', 'slideOutUp'],
		devServer: {
			proxy: proxyConfig,
			host: process.env.PUBLIC_URL ? 'localhost' : undefined,
			https: process.env.PROTOCOL === 'https://'
		}
	};
};

module.exports = {
	getConfig
};
