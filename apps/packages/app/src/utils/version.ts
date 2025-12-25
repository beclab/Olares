export class Range {
	operators = ['>=', '>', '<=', '<', '='];
	operatorIndex = -1;
	rangeVersion: Version | null = null;

	constructor(range: string) {
		this.operatorIndex = this.operators.findIndex((op) => range.startsWith(op));
		if (this.operatorIndex === -1) {
			throw new Error(`Invalid range: ${range}`);
		}

		const version = range.slice(this.operators[this.operatorIndex].length);
		this.rangeVersion = new Version(version);
	}

	satisfies(check: Version) {
		if (!this.rangeVersion) {
			throw new Error('Invalid range');
		}
		switch (this.operators[this.operatorIndex]) {
			case '>=':
				return (
					check.isGreater(this.rangeVersion) || check.isEqual(this.rangeVersion)
				);
			case '>':
				return check.isGreater(this.rangeVersion);
			case '<=':
				return (
					check.isLess(this.rangeVersion) || check.isEqual(this.rangeVersion)
				);
			case '<':
				return check.isLess(this.rangeVersion);
			case '=':
				return check.isEqual(this.rangeVersion);
			default:
				throw new Error(
					`Invalid operator: ${this.operators[this.operatorIndex]}`
				);
		}
	}
}

export class Version {
	major = 0;
	minor = 0;
	patch = 0;
	preRelease: string[] = [];

	constructor(version) {
		let mainVersion = version;
		const preReleaseIndex = version.indexOf('-');
		if (preReleaseIndex !== -1) {
			mainVersion = version.substring(0, preReleaseIndex);
			this.preRelease = version.substring(preReleaseIndex + 1).split('.');
		}

		const versionParts = mainVersion.split('.');
		if (versionParts.length !== 3) {
			throw new Error('Invalid version format');
		}
		this.major = parseInt(versionParts[0], 10);
		this.minor = parseInt(versionParts[1], 10);
		this.patch = parseInt(versionParts[2], 10);
	}

	comparePreRelease(otherPreRelease: string[]) {
		// Handle cases where one or both preRelease arrays are empty
		if (this.preRelease.length === 0 && otherPreRelease.length > 0) {
			return 1;
		} else if (this.preRelease.length > 0 && otherPreRelease.length === 0) {
			return -1;
		}
		for (
			let i = 0;
			i < Math.max(this.preRelease.length, otherPreRelease.length);
			i++
		) {
			const thisPart = this.preRelease[i] || '';
			const otherPart = otherPreRelease[i] || '';

			//Non-empty strings are considered larger
			if (thisPart === '' && otherPart !== '') {
				return -1;
			} else if (thisPart !== '' && otherPart === '') {
				return 1;
			}

			if (thisPart !== otherPart) {
				const thisIsNumber = /^\d+$/.test(thisPart);
				const otherIsNumber = /^\d+$/.test(otherPart);

				if (thisIsNumber && otherIsNumber) {
					return parseInt(thisPart, 10) - parseInt(otherPart, 10); // Compare as numbers
				} else if (thisIsNumber) {
					return -1; // Numbers are always considered lower than strings
				} else if (otherIsNumber) {
					return 1; // Strings are always considered higher than numbers
				} else {
					return thisPart.localeCompare(otherPart); // Compare as normal strings
				}
			}
		}
		return this.preRelease.length - otherPreRelease.length;
	}

	isEqual(other) {
		return (
			this.major === other.major &&
			this.minor === other.minor &&
			this.patch === other.patch &&
			this.comparePreRelease(other.preRelease) === 0
		);
	}

	isGreater(other) {
		if (this.major !== other.major) {
			return this.major > other.major;
		}
		if (this.minor !== other.minor) {
			return this.minor > other.minor;
		}
		if (this.patch !== other.patch) {
			return this.patch > other.patch;
		}
		return this.comparePreRelease(other.preRelease) > 0;
	}

	isLess(other) {
		return !this.isEqual(other) && !this.isGreater(other);
	}

	toString(): string {
		let versionStr = `${this.major}.${this.minor}.${this.patch}`;
		if (this.preRelease.length > 0) {
			versionStr += `-${this.preRelease.join('.')}`;
		}
		return versionStr;
	}
}

export function testSatisfies() {
	const versions = [
		new Version('1.0.0-0'),
		new Version('1.0.0-20140719'),
		new Version('1.0.0-alpha'),
		new Version('1.0.0-alpha.1'),
		new Version('1.0.0-alpha.beta'),
		new Version('1.0.0-beta'),
		new Version('1.0.0-beta.2'),
		new Version('1.0.0-beta.11'),
		new Version('1.0.0-rc.1'),
		new Version('1.0.0-rc.2'),
		new Version('1.0.0')
	];

	console.log('<<<<<<正向测试开始：');
	for (let i = 0; i < versions.length - 1; i++) {
		const current = versions[i];
		const next = versions[i + 1];
		const result = current.isLess(next);
		console.log(`当前OS版本：${current}`);
		console.log(`检测app版本：${next}`);
		console.log(`符合预期：${result}`);
		console.log('---');
	}

	console.log('>>>>>>逆向测试开始：');
	for (let i = versions.length - 1; i > 0; i--) {
		const current = versions[i];
		const previous = versions[i - 1];
		const result = current.isGreater(previous);
		console.log(`当前OS版本：${current}`);
		console.log(`检测app版本：${previous}`);
		console.log(`符合预期：${result}`);
		console.log('---');
	}

	console.log('======相同版本号测试开始：');
	for (let i = 0; i < versions.length; i++) {
		const current = versions[i];
		const result = current.isEqual(current);
		console.log(`当前OS版本：${current}`);
		console.log(`检测app版本：${current}`);
		console.log(`符合预期：${result}`);
		console.log('---');
	}
}
