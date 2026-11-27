import Chain from 'webpack-chain';

const { build } = require('../../config/Wise/config');

it('keeps reader fonts external without overriding shared chunk configuration', () => {
	const chain = new Chain();
	const splitChunks = { chunks: 'all', minSize: 20000 };
	chain.optimization.splitChunks(splitChunks);
	chain.module
		.rule('fonts')
		.use('url-loader')
		.loader('url-loader')
		.options({ limit: 10000 });
	build.chainWebpack({ prod: true }, chain, { isClient: true });
	expect(chain.optimization.get('splitChunks')).toBe(splitChunks);
	expect(
		chain.module.rule('fonts').use('url-loader').get('options').limit
	).toBe(false);
});
