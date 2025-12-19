const path = require('path');

const chainWebpack = (ctx, chain, { isClient }) => {
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-file-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Wizard/public/'),
					to: ''
				}
			]
		}
	]);
};

const loginConfig = {
	build: {
		chainWebpack
	}
};

module.exports = loginConfig;
