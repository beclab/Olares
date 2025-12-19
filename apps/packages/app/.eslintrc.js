module.exports = {
	root: true,
	env: {
		browser: true,
		es2021: true,
		node: true,
		webextensions: true
	},

	extends: [
		'eslint:recommended',
		'plugin:@typescript-eslint/recommended',
		'plugin:vue/vue3-essential',
		'plugin:prettier/recommended'
	],

	parserOptions: {
		ecmaVersion: 'latest',
		parser: '@typescript-eslint/parser',
		sourceType: 'module'
	},

	plugins: ['@typescript-eslint', 'vue'],

	globals: {
		NodeJS: true
	},

	rules: {
		'no-debugger': process.env.NODE_ENV === 'production' ? 'error' : 'off',
		'no-useless-escape': 0,
		'no-extra-boolean-cast': 0,
		'no-async-promise-executor': 0,
		'@typescript-eslint/no-var-requires': 0,
		'@typescript-eslint/no-explicit-any': 'off',
		'@typescript-eslint/no-non-null-assertion': 'off',
		'@typescript-eslint/no-this-alias': 'off',
		'@typescript-eslint/ban-ts-comment': 'off',
		'@typescript-eslint/no-empty-function': 'off',
		'vue/multi-word-component-names': 'off',
		'@typescript-eslint/explicit-function-return-type': 'off',
		'vue/no-unused-vars': 'off',
		'@typescript-eslint/no-unused-vars': 'off'
	}
};
