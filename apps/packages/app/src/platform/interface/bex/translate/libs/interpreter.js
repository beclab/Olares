class SimpleInterpreter {
	constructor() {
		this.exports = new Proxy(
			{},
			{
				get(target, prop) {
					return (...args) => {
						console.log(`[Interpreter] mock call: ${prop}`, args);
						if (prop === 'resHook') {
							return [args[1] || '', false];
						} else if (prop === 'reqHook') {
							return [args[3] || '', {}];
						}
						return undefined;
					};
				}
			}
		);
	}

	run(code) {
		console.log('[Interpreter] skip code execution');
	}
}

const interpreter = new SimpleInterpreter();

export default interpreter;
