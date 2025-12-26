type Callback<T> = (args?: T) => void;

interface Task<T> {
	id: string;
	callback: Callback<T>;
	args?: T;
	executed: boolean; // 是否已执行
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
	 * 设置共享检查条件
	 * @param condition 检查条件函数
	 * @param interval 检查间隔（毫秒）
	 * @param timeout 超时时间（毫秒）
	 * @param maxAttempts 最大尝试次数
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
	 * 添加任务到队列（不立即执行，等待条件满足）
	 * @param id 任务唯一标识
	 * @param callback 条件满足时的回调
	 * @param args 传递给回调的参数
	 */
	public addTask(id: string, callback: Callback<T>, args?: T): void {
		if (!this.sharedCondition) {
			throw new Error('请先设置共享检查条件');
		}

		// 如果任务已存在，先移除
		if (this.taskQueue.has(id)) {
			this.removeTask(id);
		}

		// 添加新任务
		this.taskQueue.set(id, { id, callback, args, executed: false });

		// 如果队列未运行，启动条件检查
		if (!this.running) {
			this.startChecking();
		}

		// 如果条件已满足，立即执行新添加的任务
		if (this.conditionMet) {
			this.executeTask(id);
		}
	}

	/**
	 * 批量添加任务
	 * @param tasks 任务数组
	 */
	public addTasks(
		tasks: { id: string; callback: Callback<T>; args?: T }[]
	): void {
		tasks.forEach((task) => this.addTask(task.id, task.callback, task.args));
	}

	/**
	 * 启动条件检查
	 */
	private startChecking(): void {
		if (this.running) return;

		this.running = true;
		this.attempts = 0;

		// 设置主定时器
		this.intervalId = setInterval(() => {
			this.checkCondition();
		}, this.checkInterval);

		// 设置超时定时器
		setTimeout(() => {
			if (this.running && !this.conditionMet) {
				console.warn('条件检查已超时');
				this.stopChecking();
			}
		}, this.timeout);
	}

	/**
	 * 检查共享条件
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
			console.error('条件检查出错:', error);
			this.stopChecking();
		}
	}

	/**
	 * 执行单个任务
	 */
	private executeTask(id: string): void {
		const task = this.taskQueue.get(id);
		if (task && !task.executed) {
			try {
				task.callback(task.args);
				task.executed = true;
			} catch (error) {
				console.error(`执行任务 ${id} 时出错:`, error);
			}
		}
	}

	/**
	 * 执行所有任务
	 */
	private executeAllTasks(): void {
		this.taskQueue.forEach((task) => {
			if (!task.executed) {
				this.executeTask(task.id);
			}
		});
	}

	/**
	 * 停止条件检查
	 */
	private stopChecking(): void {
		this.running = false;
		if (this.intervalId) {
			clearInterval(this.intervalId);
			this.intervalId = undefined;
		}
	}

	/**
	 * 移除任务
	 * @param id 任务唯一标识
	 */
	public removeTask(id: string): void {
		this.taskQueue.delete(id);
	}

	/**
	 * 清空所有任务
	 */
	public clearTasks(): void {
		this.taskQueue.clear();
	}

	/**
	 * 获取队列中任务数量
	 */
	public getQueueSize(): number {
		return this.taskQueue.size;
	}

	/**
	 * 检查条件是否已满足
	 */
	public hasConditionMet(): boolean {
		return this.conditionMet;
	}

	/**
	 * 重置状态，允许重新检查条件
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
