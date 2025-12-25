const path = require('path');

const chainWebpack = (ctx, chain) => {
	const CopyPlugin = require('copy-webpack-plugin');
	chain.plugin('copy-public-plugin').use(CopyPlugin, [
		{
			patterns: [
				{
					from: path.resolve('./config/Vault/public/'),
					to: ''
				}
			]
		}
	]);

	if (ctx.dev) {
		chain.plugin('copy-public-plugin').use(CopyPlugin, [
			{
				patterns: [
					{
						from: path.resolve('./config/Vault/dev/'),
						to: ''
					}
				]
			}
		]);
		return;
	}
	const wasmRoot = './dist/apps/vault/';

	const copyFileArray = [
		{
			fromPath: path.resolve(
				'./node_modules/@trustwallet/wallet-core/dist/lib/'
			),
			fromName: 'wallet-core.wasm',
			toPath: path.resolve(wasmRoot + 'js/'),
			toName: 'wallet-core.wasm'
		}
	];

	const CopyWebpackPlugin = require('../../build/plugins/CopyFilePlugin');

	chain.plugin('copy-file-plugin').use(CopyWebpackPlugin, [copyFileArray]);
};

const vaultConfig = {
	build: {
		chainWebpack
	}
};

module.exports = vaultConfig;
