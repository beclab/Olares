// const { configure } = require('quasar/wrappers');
const boot = [
	'i18n',
	'baseAxios',
	'wiseUI',
	'worker',
	'application/wise',
	'hotkeys'
];
const css = ['app.scss', 'document.scss'];
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
	if (!ctx.dev) {
		css.push('font.pro.scss');
	}

	const chainWebpack = (ctx, chain, { isClient }) => {
		const chainWebpack = require('./Wise/config');
		chainWebpack.build.chainWebpack(ctx, chain, { isClient });
	};

	return {
		boot,
		css,
		build: {
			env: {},
			chainWebpack,
			extendWebpack
			// distDir: 'dist/apps/wise'
		},
		sourceFiles: {
			indexHtmlTemplate: 'src/index.template.wise.html'
		},
		htmlVariables: {
			productName: 'Wise'
		}
	};
};

module.exports = {
	getConfig
};
