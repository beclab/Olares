const path = require('path');

const chainWebpack = (ctx, chain, { isClient }) => {
	// Fonts: never inline as data URIs. Quasar's default url-loader embeds any
	// font under 10 KB into the stylesheet as base64. The reader's Han faces are
	// sliced by unicode-range and a dozen of the rarer slices are that small;
	// inlined, they would ride along in the CSS of every reader whether or not
	// the article needs them, a third larger and uncacheable on their own.
	chain.module
		.rule('fonts')
		.use('url-loader')
		.tap((options) => ({ ...options, limit: false }));
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-file-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Wise/public/'),
					to: ''
				}
			]
		}
	]);
};

const wiseConfig = {
	build: {
		chainWebpack
	}
};

module.exports = wiseConfig;
