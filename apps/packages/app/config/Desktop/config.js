const path = require('path');

const chainWebpack = (ctx, chain, { isClient }) => {
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-file-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Desktop/public/'), // 新目录名
					to: '' // 输出到构建根目录
				}
			]
		}
	]);
};

const desktopConfig = {
	build: {
		chainWebpack
	}
};

module.exports = desktopConfig;
