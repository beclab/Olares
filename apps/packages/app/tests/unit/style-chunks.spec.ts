import { mkdtempSync, mkdirSync, writeFileSync, readFileSync } from 'fs';
import { tmpdir } from 'os';
import { join } from 'path';
import webpack from 'webpack';
import Chain from 'webpack-chain';

const { rmSync } = require('fs');
const MiniCssExtractPlugin = require('mini-css-extract-plugin');
const { splitStyleChunks } = require('../../build/webpack/styleChunks');

const chunks = (chunk: { name: string }) =>
	!/sharedWebsocketWorker/i.test(chunk.name || '');
const original = {
	chunks,
	minSize: 0,
	cacheGroups: {
		defaultVendors: {
			test: /[\\/]node_modules[\\/]/,
			priority: -10,
			reuseExistingChunk: true
		},
		default: { minChunks: 2, priority: -20, reuseExistingChunk: true }
	}
};

it('preserves worker isolation and resolves scoped/nested vendor stylesheet names', () => {
	const options = splitStyleChunks(original);
	expect(options.chunks).toBe(chunks);
	expect(options.chunks({ name: 'sharedWebsocketWorker' })).toBe(false);
	const name = options.cacheGroups.vendorStyles.name;
	expect(
		name({ nameForCondition: () => '/app/node_modules/@vue-office/docx/a.css' })
	).toBe('vendor-css-vue-office--docx');
	expect(
		name({
			nameForCondition: () =>
				'C:\\app\\node_modules\\parent\\node_modules\\video.js\\a.css'
		})
	).toBe('vendor-css-video.js');
});

it('applies shared style splitting to production clients and preserves BEX/development behavior', () => {
	const application = process.env.APPLICATION;
	process.env.APPLICATION = 'SHARE';
	try {
		for (const [prod, modeName, isClient] of [
			[true, 'spa', true],
			[false, 'spa', true],
			[true, 'bex', true],
			[true, 'spa', false]
		] as const) {
			const chain = new Chain();
			const config = require('../../config')({ prod, dev: !prod, modeName });
			config.build.chainWebpack(chain, { isClient });
			const options = chain.optimization.get('splitChunks');
			if (prod && modeName === 'spa' && isClient) {
				expect(options.cacheGroups.vendorStyles.type).toBe('css/mini-extract');
				expect(
					options.cacheGroups.default.test({ type: 'css/mini-extract' })
				).toBe(false);
				expect(options.chunks({ name: 'sharedWebsocketWorker' })).toBe(false);
			} else {
				expect(options?.cacheGroups.vendorStyles).toBeUndefined();
			}
		}
	} finally {
		if (application === undefined) delete process.env.APPLICATION;
		else process.env.APPLICATION = application;
	}
});

it('builds opposite lazy-route imports without order warnings or duplicated vendor CSS', async () => {
	const directory = mkdtempSync(join(tmpdir(), 'termipass-style-chunks-'));
	const write = (file: string, content: string) => {
		const target = join(directory, file);
		mkdirSync(join(target, '..'), { recursive: true });
		writeFileSync(target, content);
	};
	write(
		'index.js',
		`import 'library-a/style.css';
import(/* webpackChunkName: "route-a" */ './a.js');
import(/* webpackChunkName: "route-b" */ './b.js');`
	);
	write(
		'a.js',
		`import './app-a.css'; import './app-b.css';
import 'library-a/style.css'; import 'library-b/style.css';`
	);
	write(
		'b.js',
		`import './app-b.css'; import './app-a.css';
import 'library-b/style.css'; import 'library-a/style.css';`
	);
	write('app-a.css', '.app-a { color: red; }');
	write('app-b.css', '.app-b { color: blue; }');
	write('node_modules/library-a/style.css', '.vendor-a { color: red; }');
	write('node_modules/library-b/style.css', '.vendor-b { color: blue; }');
	const compile = (splitChunks: any, output: string) =>
		new Promise<webpack.Stats>((resolve, reject) => {
			const compiler = webpack({
				mode: 'production',
				context: directory,
				entry: './index.js',
				output: { path: join(directory, output), filename: '[name].js' },
				module: {
					rules: [
						{
							test: /\.css$/,
							use: [MiniCssExtractPlugin.loader, require.resolve('css-loader')]
						}
					]
				},
				plugins: [new MiniCssExtractPlugin({ filename: '[name].css' })],
				optimization: { minimize: false, splitChunks }
			});
			compiler.run((error, stats) => {
				compiler.close((closeError) => {
					if (error || closeError || !stats) reject(error || closeError);
					else resolve(stats);
				});
			});
		});
	try {
		const before = await compile(original, 'before');
		expect(before.hasErrors()).toBe(false);
		expect(
			before
				.toJson()
				.warnings?.some((warning) => /Conflicting order/.test(warning.message))
		).toBe(true);
		const after = await compile(splitStyleChunks(original), 'after');
		expect(after.toJson().errors).toEqual([]);
		expect(after.toJson().warnings).toEqual([]);
		const css = after.compilation
			.getAssets()
			.filter((asset) => asset.name.endsWith('.css'));
		expect(css.map((asset) => asset.name)).toEqual(
			expect.arrayContaining([
				'route-a.css',
				'route-b.css',
				'vendor-css-library-a.css',
				'vendor-css-library-b.css'
			])
		);
		const initialFiles = after.compilation.entrypoints.get('main')!.getFiles();
		expect(initialFiles).toContain('vendor-css-library-a.css');
		expect(initialFiles).not.toContain('vendor-css-library-b.css');
		const contents = css
			.map((asset) =>
				readFileSync(join(directory, 'after', asset.name), 'utf8')
			)
			.join('\n');
		expect(contents.match(/\.vendor-a/g)).toHaveLength(1);
		expect(contents.match(/\.vendor-b/g)).toHaveLength(1);
		expect(contents.match(/\.app-a/g)).toHaveLength(2);
	} finally {
		rmSync(directory, { recursive: true, force: true });
	}
}, 20000);
