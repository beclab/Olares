const path = require('path');

const chainWebpack = (ctx, chain, { isClient }) => {
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-file-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Settings/public/'),
					to: ''
				}
			]
		}
	]);
};

const settingsConfig = {
	build: {
		chainWebpack
	}
};

module.exports = settingsConfig;
