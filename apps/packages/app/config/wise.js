// const { configure } = require('quasar/wrappers');
const boot = [
	'i18n',
	'baseAxios',
	'wiseUI',
	'worker',
	'application/wise',
	'hotkeys'
];
const css = ['wise/app.scss'];
const path = require('path');

const extendWebpack = (ctx, cfg) => {
	cfg.module.rules.push({
		test: /\.worker\.(ts|js)$/,
		include: path.resolve(__dirname, 'src'),
		exclude: /node_modules/,
		use: {
			loader: 'worker-loader',
			options: {
				inline: 'fallback',
				filename: '[name].[contenthash].js',
				crossOrigin: 'anonymous'
			}
		}
	});
};
const getConfig = (ctx) => {
	const proxy = { ...require('./proxyDefault') };
	// /download is a Wise page route; only forward its API prefix so opening or
	// refreshing the page serves the SPA.
	proxy['/download/api'] = proxy['/download'];
	delete proxy['/download'];

	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpack = require('./Wise/config');
		chainWebpack.build.chainWebpack(ctx, chain, { isClient });
	};

	return {
		boot,
		devServer: {
			proxy: {
				// Local dev proxy for Files raw and preview endpoints.
				'/api/raw/': {
					target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
					changeOrigin: true
				},
				'/api/preview/': {
					target: `${process.env.PROTOCOL}files.${process.env.ACCOUNT_DOMAIN}`,
					changeOrigin: true
				},
				...proxy
			}
		},
		css,
		build: {
			env: {},
			chainWebpack,
			extendWebpack
			// distDir: 'dist/apps/wise'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.wise.html',
			variables: 'wise/variables.scss'
		},
		htmlVariables: {
			productName: 'Wise'
		}
	};
};

module.exports = {
	getConfig
};
