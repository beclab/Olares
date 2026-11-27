/* eslint-disable */
// https://github.com/michael-ciniawsky/postcss-load-config

module.exports = {
	plugins: [
		// Insert classic vh/vw fallbacks before dynamic viewport units (dvh/svh/lvh)
		// so layouts still render on WebView/Chromium < 108 which drop unknown units.
		// Package ships an ESM default export, hence `.default` under CommonJS.
		require('postcss-viewport-fallback').default(),
		// to edit target browsers: use "browserslist" field in package.json
		require('autoprefixer')
	]
};
