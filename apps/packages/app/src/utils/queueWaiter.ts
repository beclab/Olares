type Callback<T> = (args?: T) => void;

interface Condition<T> {
	id: string;
	condition: () => boolean;
	callback: Callback<T>;
	args?: T;
}

interface WaitTask<T> extends Condition<T> {
	timeoutId?: NodeJS.Timeout;
	attempts: number;
	maxAttempts: number;
	interval: number;
}

class Waiter<T = void> {
	private conditions: Map<string, Condition<T>> = new Map();
	private taskQueue: Map<string, WaitTask<T>> = new Map();
	private intervalId?: NodeJS.Timeout;
	private running = false;
	private checkInterval = 300;

	/**
	 * Set a check condition (without executing).
	 * @param id unique condition id
	 * @param condition condition check function
	 * @param callback callback when condition is met
	 * @param args args passed to callback
	 */
	public setCondition(
		id: string,
		condition: () => boolean,
		callback: Callback<T>,
		args?: T
	): void {
		this.conditions.set(id, { id, condition, callback, args });
	}

	/**
	 * Get an existing condition.
	 * @param id unique condition id
	 * @returns condition object or undefined
	 */
	public getCondition(id: string): Condition<T> | undefined {
		return this.conditions.get(id);
	}

	/**
	 * Create task from existing condition and enqueue it.
	 * @param id unique condition id
	 * @param interval check interval in milliseconds
	 * @param timeout timeout in milliseconds
	 * @param maxAttempts max number of attempts
	 * @returns whether task was enqueued successfully
	 */
	public enqueueTask(
		id: string,
		interval = 300,
		timeout = 10000,
		maxAttempts = Infinity
	): boolean {
		const condition = this.conditions.get(id);
		if (!condition) {
			console.warn(`条件 ${id} 不存在`);
			return false;
		}

		// If task already exists, remove it first.
		if (this.taskQueue.has(id)) {
			this.removeTask(id);
		}

		// Create task.
		const task: WaitTask<T> = {
			...condition,
			attempts: 0,
			maxAttempts,
			interval
		};

		this.taskQueue.set(id, task);

		// Start processing if queue is not running.
		if (!this.running) {
			this.startProcessing(interval);
		}

		// Set timeout.
		if (timeout > 0) {
			task.timeoutId = setTimeout(() => {
				if (this.taskQueue.has(id)) {
					console.warn(`任务 ${id} 已超时`);
					this.removeTask(id);
				}
			}, timeout);
		}

		return true;
	}

	/**
	 * Add task directly (legacy-compatible).
	 * @param id unique task id
	 * @param condition condition check function
	 * @param callback callback when condition is met
	 * @param interval check interval in milliseconds
	 * @param args args passed to callback
	 * @param timeout timeout in milliseconds
	 * @param maxAttempts max number of attempts
	 */
	public addTask(
		id: string,
		condition: () => boolean,
		callback: Callback<T>,
		interval = 300,
		args?: T,
		timeout = 10000,
		maxAttempts = Infinity
	): void {
		// Set condition first.
		this.setCondition(id, condition, callback, args);

		// Then enqueue it.
		this.enqueueTask(id, interval, timeout, maxAttempts);
	}

	/**
	 * Start queue processing.
	 */
	private startProcessing(interval: number): void {
		if (this.running) return;

		this.checkInterval = interval;
		this.running = true;
		this.intervalId = setInterval(() => {
			this.processQueue();
		}, this.checkInterval);
	}

	/**
	 * Stop queue processing.
	 */
	private stopProcessing(): void {
		if (!this.running) return;

		this.running = false;
		if (this.intervalId) {
			clearInterval(this.intervalId);
			this.intervalId = undefined;
		}
	}

	/**
	 * Process all tasks in queue.
	 */
	private processQueue(): void {
		if (this.taskQueue.size === 0) return;

		[...this.taskQueue.values()].forEach((task) => {
			try {
				task.attempts++;

				// Check whether condition is met.
				if (task.condition()) {
					task.callback(task.args);
					this.removeTask(task.id);
					return;
				}

				// Check whether max attempts has been reached.
				if (task.attempts >= task.maxAttempts) {
					console.warn(`任务 ${task.id} 已达到最大尝试次数`);
					this.removeTask(task.id);
				}
			} catch (error) {
				console.error(`处理任务 ${task.id} 时出错:`, error);
				this.removeTask(task.id);
			}
		});
	}

	/**
	 * Remove task.
	 * @param id unique task id
	 */
	public removeTask(id: string): void {
		const task = this.taskQueue.get(id);
		if (task && task.timeoutId) {
			clearTimeout(task.timeoutId);
		}
		this.taskQueue.delete(id);

		// Stop processing when queue is empty.
		if (this.taskQueue.size === 0) {
			this.stopProcessing();
		}
	}

	/**
	 * Remove existing condition (if not enqueued).
	 * @param id unique condition id
	 * @returns whether removal succeeded
	 */
	public removeCondition(id: string): boolean {
		if (this.taskQueue.has(id)) {
			console.warn(`条件 ${id} 已在队列中执行，无法移除`);
			return false;
		}

		return this.conditions.delete(id);
	}

	/**
	 * Clear all conditions and tasks.
	 */
	public clearAll(): void {
		// Clear task queue.
		this.taskQueue.forEach((task) => {
			if (task.timeoutId) {
				clearTimeout(task.timeoutId);
			}
		});
		this.taskQueue.clear();

		// Clear conditions.
		this.conditions.clear();

		// Stop processing.
		this.stopProcessing();
	}

	/**
	 * Get queue size.
	 */
	public getQueueSize(): number {
		return this.taskQueue.size;
	}

	/**
	 * Get number of conditions.
	 */
	public getConditionCount(): number {
		return this.conditions.size;
	}

	/**
	 * Get all condition IDs.
	 */
	public getAllConditionIds(): string[] {
		return [...this.conditions.keys()];
	}

	/**
	 * Get all task IDs in queue.
	 */
	public getQueueTaskIds(): string[] {
		return [...this.taskQueue.keys()];
	}

	/**
	 * Check if condition exists.
	 * @param id unique condition id
	 */
	public hasCondition(id: string): boolean {
		return this.conditions.has(id);
	}

	/**
	 * Check if task is in queue.
	 * @param id unique task id
	 */
	public hasTask(id: string): boolean {
		return this.taskQueue.has(id);
	}
}

export default Waiter;
