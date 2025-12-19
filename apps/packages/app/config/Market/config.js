const path = require('path');

const chainWebpack = (ctx, chain, { isClient }) => {
	const updatePackageVersionByLatestTag = require('../../build/plugins/UpdatePackageVersionByLatestTag');
	chain
		.plugin('update-package-json')
		.use(updatePackageVersionByLatestTag, [[]]);
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-file-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Market/public/'),
					to: ''
				}
			]
		}
	]);
};

const marketConfig = {
	build: {
		chainWebpack
	}
};

module.exports = marketConfig;
