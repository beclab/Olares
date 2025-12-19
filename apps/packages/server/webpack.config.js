/* eslint-disable @typescript-eslint/no-var-requires */
const path = require('path');
const AddAssetPlugin = require('add-asset-webpack-plugin');
const package = require('./package.json');
const webpack = require('webpack');
const TerserPlugin = require('terser-webpack-plugin');

const isProduction = process.env.NODE_ENV == 'production';

const config = {
	target: 'node',
	entry: './src/init.ts',
	output: {
		path: path.resolve(__dirname, 'dist')
	},
	resolve: {
		alias: {
			'@didvault': path.resolve(__dirname, '../../packages')
		}
	},
	externals: {
		bcrypt: 'commonjs2 bcrypt',
		level: 'commonjs2 level'
	},
	optimization: {
		splitChunks: {
			chunks: 'all',
			cacheGroups: {
				vendor: {
					name: 'vendor',
					priority: 1,
					test: /node_modules/,
					minSize: 0,
					minChunks: 1
				},

				common: {
					name: 'common',
					priority: 0,
					minSize: 0,
					minChunks: 2
				}
			}
		},
		minimizer: [
			new TerserPlugin({
				terserOptions: {
					mangle: false,
					keep_classnames: true,
					keep_fnames: true
				}
			})
		]
	},
	plugins: [
		new AddAssetPlugin('./package.json', createPackage),
		new webpack.IgnorePlugin({ resourceRegExp: /^pg-native$/ })
		// Or, for WebPack 4+:
	],
	module: {
		rules: [
			{
				test: /\.(ts|tsx)$/i,
				loader: 'ts-loader',
				exclude: ['/node_modules/'],
				options: {
					transpileOnly: true
				}
			}
		]
	},
	// eslint-disable-next-line no-dupe-keys
	resolve: {
		extensions: ['.tsx', '.ts', '.jsx', '.js', '.json', '...']
	}
};

module.exports = () => {
	if (isProduction) {
		config.mode = 'production';
	} else {
		config.mode = 'development';
	}
	return config;
};

function createPackage() {
	const externals = config.externals;
	const externalsKeys = Object.keys(externals);
	const dependencies = package.dependencies;
	const externals_dependencies = {};

	for (const key in dependencies) {
		if (externalsKeys.includes(key)) {
			externals_dependencies[key] = dependencies[key];
		}
	}

	const packages = {
		dependencies: externals_dependencies,
		scripts: {
			server: 'node main.js'
		}
	};
	return JSON.stringify(packages);
}
