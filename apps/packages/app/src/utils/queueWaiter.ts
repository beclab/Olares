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
	 * 设置检查条件（不执行）
	 * @param id 条件唯一标识
	 * @param condition 检查条件函数
	 * @param callback 条件满足时的回调
	 * @param args 传递给回调的参数
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
	 * 获取已设置的条件
	 * @param id 条件唯一标识
	 * @returns 条件对象或 undefined
	 */
	public getCondition(id: string): Condition<T> | undefined {
		return this.conditions.get(id);
	}

	/**
	 * 从已设置的条件创建任务并添加到队列执行
	 * @param id 条件唯一标识
	 * @param interval 检查间隔（毫秒）
	 * @param timeout 超时时间（毫秒）
	 * @param maxAttempts 最大尝试次数
	 * @returns 是否成功添加任务
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

		// 如果任务已存在，先移除
		if (this.taskQueue.has(id)) {
			this.removeTask(id);
		}

		// 创建任务
		const task: WaitTask<T> = {
			...condition,
			attempts: 0,
			maxAttempts,
			interval
		};

		this.taskQueue.set(id, task);

		// 如果队列未运行，启动队列处理
		if (!this.running) {
			this.startProcessing(interval);
		}

		// 设置超时
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
	 * 直接添加任务（兼容旧用法）
	 * @param id 任务唯一标识
	 * @param condition 检查条件函数
	 * @param callback 条件满足时的回调
	 * @param interval 检查间隔（毫秒）
	 * @param args 传递给回调的参数
	 * @param timeout 超时时间（毫秒）
	 * @param maxAttempts 最大尝试次数
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
		// 先设置条件
		this.setCondition(id, condition, callback, args);

		// 再添加到队列
		this.enqueueTask(id, interval, timeout, maxAttempts);
	}

	/**
	 * 启动队列处理
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
	 * 停止队列处理
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
	 * 处理队列中的所有任务
	 */
	private processQueue(): void {
		if (this.taskQueue.size === 0) return;

		[...this.taskQueue.values()].forEach((task) => {
			try {
				task.attempts++;

				// 检查条件是否满足
				if (task.condition()) {
					task.callback(task.args);
					this.removeTask(task.id);
					return;
				}

				// 检查是否超过最大尝试次数
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
	 * 移除任务
	 * @param id 任务唯一标识
	 */
	public removeTask(id: string): void {
		const task = this.taskQueue.get(id);
		if (task && task.timeoutId) {
			clearTimeout(task.timeoutId);
		}
		this.taskQueue.delete(id);

		// 如果队列空了，停止处理
		if (this.taskQueue.size === 0) {
			this.stopProcessing();
		}
	}

	/**
	 * 移除已设置的条件（如果未入队）
	 * @param id 条件唯一标识
	 * @returns 是否成功移除
	 */
	public removeCondition(id: string): boolean {
		if (this.taskQueue.has(id)) {
			console.warn(`条件 ${id} 已在队列中执行，无法移除`);
			return false;
		}

		return this.conditions.delete(id);
	}

	/**
	 * 清空所有条件和任务
	 */
	public clearAll(): void {
		// 清理任务队列
		this.taskQueue.forEach((task) => {
			if (task.timeoutId) {
				clearTimeout(task.timeoutId);
			}
		});
		this.taskQueue.clear();

		// 清理条件
		this.conditions.clear();

		// 停止处理
		this.stopProcessing();
	}

	/**
	 * 获取队列中任务数量
	 */
	public getQueueSize(): number {
		return this.taskQueue.size;
	}

	/**
	 * 获取已设置的条件数量
	 */
	public getConditionCount(): number {
		return this.conditions.size;
	}

	/**
	 * 获取所有条件的ID
	 */
	public getAllConditionIds(): string[] {
		return [...this.conditions.keys()];
	}

	/**
	 * 获取队列中所有任务的ID
	 */
	public getQueueTaskIds(): string[] {
		return [...this.taskQueue.keys()];
	}

	/**
	 * 检查条件是否存在
	 * @param id 条件唯一标识
	 */
	public hasCondition(id: string): boolean {
		return this.conditions.has(id);
	}

	/**
	 * 检查任务是否在队列中
	 * @param id 任务唯一标识
	 */
	public hasTask(id: string): boolean {
		return this.taskQueue.has(id);
	}
}

export default Waiter;
