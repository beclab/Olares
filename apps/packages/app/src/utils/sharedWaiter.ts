type Callback<T> = (args?: T) => void;

interface Task<T> {
	id: string;
	callback: Callback<T>;
	args?: T;
	executed: boolean; // Whether already executed
}

class Waiter<T = void> {
	private sharedCondition?: () => boolean;
	private taskQueue: Map<string, Task<T>> = new Map();
	private checkInterval = 300;
	private timeout = 10000;
	private maxAttempts = Infinity;
	private intervalId?: NodeJS.Timeout;
	private attempts = 0;
	private running = false;
	private conditionMet = false;

	/**
	 * Set shared check condition.
	 * @param condition condition check function
	 * @param interval check interval in milliseconds
	 * @param timeout timeout in milliseconds
	 * @param maxAttempts max number of attempts
	 */
	public setCondition(
		condition: () => boolean,
		interval = 300,
		timeout = 10000,
		maxAttempts = Infinity
	): void {
		this.sharedCondition = condition;
		this.checkInterval = interval;
		this.timeout = timeout;
		this.maxAttempts = maxAttempts;
	}

	/**
	 * Add task to queue (wait until condition is met).
	 * @param id unique task id
	 * @param callback callback when condition is met
	 * @param args args passed to callback
	 */
	public addTask(id: string, callback: Callback<T>, args?: T): void {
		if (!this.sharedCondition) {
			throw new Error('请先设置共享检查条件');
		}

		// If task already exists, remove it first.
		if (this.taskQueue.has(id)) {
			this.removeTask(id);
		}

		// Add new task.
		this.taskQueue.set(id, { id, callback, args, executed: false });

		// Start checking if queue is not running.
		if (!this.running) {
			this.startChecking();
		}

		// Run newly added task immediately if condition is already met.
		if (this.conditionMet) {
			this.executeTask(id);
		}
	}

	/**
	 * Add tasks in batch.
	 * @param tasks task list
	 */
	public addTasks(
		tasks: { id: string; callback: Callback<T>; args?: T }[]
	): void {
		tasks.forEach((task) => this.addTask(task.id, task.callback, task.args));
	}

	/**
	 * Start condition checking.
	 */
	private startChecking(): void {
		if (this.running) return;

		this.running = true;
		this.attempts = 0;

		// Start main timer.
		this.intervalId = setInterval(() => {
			this.checkCondition();
		}, this.checkInterval);

		// Start timeout timer.
		setTimeout(() => {
			if (this.running && !this.conditionMet) {
				console.warn('条件检查已超时');
				this.stopChecking();
			}
		}, this.timeout);
	}

	/**
	 * Check shared condition.
	 */
	private checkCondition(): void {
		if (this.conditionMet || this.attempts >= this.maxAttempts) {
			this.stopChecking();
			return;
		}

		this.attempts++;

		try {
			if (this.sharedCondition!()) {
				this.conditionMet = true;
				this.executeAllTasks();
				this.stopChecking();
			}
		} catch (error) {
			console.error('Condition check failed:', error);
			this.stopChecking();
		}
	}

	/**
	 * Execute a single task.
	 */
	private executeTask(id: string): void {
		const task = this.taskQueue.get(id);
		if (task && !task.executed) {
			try {
				task.callback(task.args);
				task.executed = true;
			} catch (error) {
				console.error(`Failed to execute task ${id}:`, error);
			}
		}
	}

	/**
	 * Execute all tasks.
	 */
	private executeAllTasks(): void {
		this.taskQueue.forEach((task) => {
			if (!task.executed) {
				this.executeTask(task.id);
			}
		});
	}

	/**
	 * Stop condition checking.
	 */
	private stopChecking(): void {
		this.running = false;
		if (this.intervalId) {
			clearInterval(this.intervalId);
			this.intervalId = undefined;
		}
	}

	/**
	 * Remove task.
	 * @param id unique task id
	 */
	public removeTask(id: string): void {
		this.taskQueue.delete(id);
	}

	/**
	 * Clear all tasks.
	 */
	public clearTasks(): void {
		this.taskQueue.clear();
	}

	/**
	 * Get queue size.
	 */
	public getQueueSize(): number {
		return this.taskQueue.size;
	}

	/**
	 * Check whether condition has been met.
	 */
	public hasConditionMet(): boolean {
		return this.conditionMet;
	}

	/**
	 * Reset state and allow condition re-check.
	 */
	public reset(): void {
		this.stopChecking();
		this.conditionMet = false;
		this.attempts = 0;
		this.taskQueue.forEach((task) => {
			task.executed = false;
		});
	}
}

export default Waiter;
