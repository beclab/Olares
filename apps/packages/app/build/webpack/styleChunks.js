const CSS_MODULE_TYPE = 'css/mini-extract';
const NODE_MODULES = /[\\/]node_modules[\\/]/;

function vendorStyleChunkName(module) {
	const resource = module.nameForCondition();
	const parts = resource.split(NODE_MODULES).pop().split(/[\\/]/);
	const packageName = parts[0].startsWith('@')
		? `${parts[0].slice(1)}--${parts[1]}`
		: parts[0];
	return `vendor-css-${packageName}`;
}

// Application CSS stays with the route that imports it. Each third-party
// stylesheet family is shared separately, so unrelated lazy viewers cannot
// impose contradictory ordering on one large vendor stylesheet.
function splitStyleChunks(splitChunks) {
	return {
		...splitChunks,
		cacheGroups: {
			...splitChunks.cacheGroups,
			defaultVendors: {
				...splitChunks.cacheGroups.defaultVendors,
				test: (module) =>
					module.type !== CSS_MODULE_TYPE &&
					NODE_MODULES.test(module.nameForCondition() || '')
			},
			default: {
				...splitChunks.cacheGroups.default,
				test: (module) => module.type !== CSS_MODULE_TYPE
			},
			vendorStyles: {
				type: CSS_MODULE_TYPE,
				test: NODE_MODULES,
				name: vendorStyleChunkName,
				priority: 10,
				enforce: true,
				reuseExistingChunk: true
			}
		}
	};
}

module.exports = { splitStyleChunks };
