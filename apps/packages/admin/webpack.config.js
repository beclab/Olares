/* eslint-disable @typescript-eslint/no-var-requires */
const path = require('path');
const webpack = require('webpack');
const nodeExternals = require('webpack-node-externals');
const ForkTsCheckerWebpackPlugin = require('fork-ts-checker-webpack-plugin');
const AddAssetPlugin = require('add-asset-webpack-plugin');
const package = require('./package.json');

const isProduction = process.env.NODE_ENV == 'production';

const config = {
  entry: './src/main',
  target: 'node',

  externals: {
    level: 'commonjs2 level',
    bcrypt: 'commonjs2 bcrypt',
  },
  // resolve: {
  //   extensions: ['.js', '.ts', '.json'],
  // },
  resolve: {
    extensions: ['.js', '.ts', '.json'],
  },

  module: {
    rules: [
      {
        test: /\.ts?$/,
        use: {
          loader: 'ts-loader',
          options: { transpileOnly: true },
        },
        exclude: /node_modules/,
      },
    ],
  },

  output: {
    filename: 'main.js',
    path: path.resolve(__dirname, 'dist'),
  },

  plugins: [
    new webpack.IgnorePlugin({
      checkResource(resource) {
        const lazyImports = [
          '@nestjs/microservices',
          '@nestjs/microservices/microservices-module',
          '@nestjs/websockets/socket-module',
          'cache-manager',
          'class-validator',
          'class-transformer',
          'pg-native',
        ];
        if (!lazyImports.includes(resource)) {
          return false;
        }
        try {
          require.resolve(resource, {
            paths: [process.cwd()],
          });
        } catch (err) {
          return true;
        }
        return false;
      },
    }),
    new ForkTsCheckerWebpackPlugin(),
    new AddAssetPlugin('./package.json', createPackage),
    //  new webpack.IgnorePlugin({ resourceRegExp: /^pg-native$/ }),
  ],
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
      server: 'node main.js',
    },
  };
  return JSON.stringify(packages);
}

module.exports = () => {
  if (isProduction) {
    config.mode = 'production';
  } else {
    config.mode = 'development';
  }
  return config;
};
