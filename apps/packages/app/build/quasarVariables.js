const path = require('path');
const fs = require('fs');

function changeVariablesCssSource(config) {
	const quasarVariables = 'quasar.variables.scss';
	const variablesConfigFile = config.sourceFiles.variables;
	let variablesFile = '';
	if (config.sourceFiles && config.sourceFiles.variables) {
		const variablesFilePath = path.resolve(
			__dirname,
			'../src/css',
			variablesConfigFile
		);
		if (!fs.existsSync(variablesFilePath)) {
			console.error(
				'\x1b[31m%s\x1b[0m',
				`[Configuration Error] Variables file specified in config.sourceFiles.variables does not exist:

				Full path: ${variablesFilePath}
				Current config value: ${variablesConfigFile}

				Possible solutions:
				1. Check if the filename is correct and the file exists in the src/ directory
				2. If you don't need a custom variables file, remove config.sourceFiles.variables
				3. Verify the file path uses correct path separators
				`
			);
			process.exit(1);
			return;
		} else {
			variablesFile = variablesConfigFile;
		}
	} else {
		variablesFile = quasarVariables;
	}

	const variablesFileParser = path.parse(variablesFile);
	const variablesFileName = variablesFile.replace(variablesFileParser.ext, '');

	changeQuasarLoaderVariablesFile(variablesFileName, 'scss');
	changeQuasarLoaderVariablesFile(variablesFileName, 'sass');
	changeQuasarHelperVariablesFile(variablesFileName);
}

function changeQuasarLoaderVariablesFile(replaceContent, variablesFileExt) {
	const controlHubVariablesPath = `src/css/${replaceContent}`;
	const variablesScssLoaderPath = path.join(
		__dirname,
		`../node_modules/@quasar/app-webpack/lib/webpack/loader.quasar-${variablesFileExt}-variables.js`
	);

	updateFileContent(
		variablesScssLoaderPath,
		/(?<=~src\/css\/).*(?=\.\$\{)/,
		replaceContent
	);
}

function changeQuasarHelperVariablesFile(replaceContent) {
	const variablesScssLoaderPath = path.join(
		__dirname,
		`../node_modules/@quasar/app-webpack/lib/helpers/css-variables.js`
	);

	updateFileContent(
		variablesScssLoaderPath,
		/(?<=css\/).*(?=\.scss)/,
		replaceContent
	);
}

function updateFileContent(
	variablesScssLoaderPath,
	quasarVariablesPattern,
	replaceContent
) {
	try {
		let loaderContent = fs.readFileSync(variablesScssLoaderPath, 'utf8');
		loaderContent = loaderContent.replace(
			quasarVariablesPattern,
			replaceContent
		);
		fs.writeFileSync(variablesScssLoaderPath, loaderContent, 'utf8');
	} catch (error) {
		console.error('Failed to modify Quasar SCSS variables loader:', error);
	}
}

module.exports = changeVariablesCssSource;
