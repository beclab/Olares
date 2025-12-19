const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

module.exports = class UpdatePackageVersionByLatestTag {
	constructor(paths) {
		this.paths = paths;
	}
	apply(compiler) {
		compiler.hooks.environment.tap(
			'UpdatePackageVersionByLatestTagPlugin',
			() => {
				let latestTag = '';
				try {
					execSync('git fetch -a');

					const lastCommit = execSync('git rev-list --tags --max-count=1')
						.toString()
						.replace('\n', '');

					latestTag = execSync(`git describe --tags ${lastCommit}`)
						.toString()
						.replace('\n', '');
				} catch (error) {
					console.log('get latestTag error');
					console.log(error);
					return;
				}
				const regex = /^v(\d+)\.(\d+)\.(\d+)$/;
				const match = latestTag.match(regex);
				if (!match) {
					return;
				}
				const [, major, minor, patch] = match;

				try {
					const appPackageJsonUrl = path.join(
						process.cwd(),
						this.paths && this.paths.length > 0 ? this.paths[0] : 'package.json'
					);
					const appPackageJson = fs.readFileSync(appPackageJsonUrl, 'utf-8');
					const appPackage = JSON.parse(appPackageJson);
					appPackage.version = `${major}.${minor}.${patch}`;
					fs.writeFileSync(
						appPackageJsonUrl,
						JSON.stringify(appPackage, null, 2)
					);
				} catch (error) {
					console.log(error);
				}
			}
		);
	}
};
