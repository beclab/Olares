import { kissLog } from './log';

export const taskPool = (
	fn,
	preFn,
	_interval = 100,
	_limit = 100,
	_retryInterval = 1000
) => {
	const pool = [];
	const maxRetry = 2;
	let maxCount = _limit;
	let curCount = 0;
	let interval = _interval;
	let lastExecutionTime = 0;
	let schedulerTimer = null;

	const scheduleNext = () => {
		if (schedulerTimer) {
			return;
		}

		if (curCount >= maxCount || pool.length === 0) {
			return;
		}

		const now = Date.now();
		const timeSinceLast = now - lastExecutionTime;
		const delay = Math.max(0, interval - timeSinceLast);

		schedulerTimer = setTimeout(() => {
			schedulerTimer = null;
			if (curCount < maxCount && pool.length > 0) {
				const task = pool.shift();
				if (task) {
					lastExecutionTime = Date.now();
					execute(task);
				}
			}

			if (pool.length > 0) {
				scheduleNext();
			}
		}, delay);
	};

	const execute = async (task) => {
		curCount++;
		const { args, resolve, reject, retry } = task;

		try {
			const preArgs = preFn ? await preFn(args) : {};
			const res = await fn({ ...args, ...preArgs });
			resolve(res);
		} catch (err) {
			kissLog(err, 'task');
			if (retry < maxRetry) {
				setTimeout(() => {
					pool.unshift({ ...task, retry: retry + 1 });
					scheduleNext();
				}, _retryInterval);
			} else {
				reject(err);
			}
		} finally {
			curCount--;
			scheduleNext();
		}
	};

	return {
		push: async (args) => {
			return new Promise((resolve, reject) => {
				pool.push({ args, resolve, reject, retry: 0 });
				scheduleNext();
			});
		},
		update: (_interval = 100, _limit = 100) => {
			if (_interval >= 0 && _interval <= 5000 && _interval !== interval) {
				interval = _interval;
			}
			if (_limit >= 1 && _limit <= 100 && _limit !== maxCount) {
				maxCount = _limit;
			}
			scheduleNext();
		},
		clear: () => {
			for (const task of pool) {
				task.reject(new Error('Task pool was cleared'));
			}
			pool.length = 0;
			curCount = 0;
			if (schedulerTimer) {
				clearTimeout(schedulerTimer);
				schedulerTimer = null;
			}
		}
	};
};
