import { AppEntry } from 'src/constant/constants';

export interface OnUpdateUITask {
	onInit(): void;
	onUpdate(app: AppEntry): void;
}

export class TaskHandler {
	private taskList: OnUpdateUITask[] = [];
	private app?: AppEntry;
	private isReady = false;
	private running = false;
	private executed = false;

	addTask(task: OnUpdateUITask, isReady = false): this {
		if (this.running) {
			console.log('task is already running');
			return this;
		}

		this.isReady = isReady;

		if (!this.taskList.includes(task)) {
			this.taskList.push(task);
		}

		if (this.isReady) {
			this.execute();
		}

		return this;
	}

	withApp(app: AppEntry) {
		if (!app) {
			console.error('app empty');
			return this;
		}

		this.app = app;

		this.execute();
	}

	private execute(): void {
		if (this.running || this.executed || !this.isReady) {
			return;
		}

		if (!this.app) {
			console.log('need app，can not execute');
			return;
		}

		if (this.taskList.length === 0) {
			console.log('no task to execute');
			return;
		}

		this.running = true;

		try {
			this.taskList.forEach((task) => {
				task.onInit();
				task.onUpdate(this.app!);
			});
			console.log(
				`${this.app.name} running success ${this.taskList.length} tasks`
			);
			this.executed = true;
		} catch (error) {
			console.error('running failure:', error);
		} finally {
			this.running = false;
		}
	}

	reset(): void {
		this.taskList = [];
		this.app = undefined;
		this.running = false;
		this.executed = false;
		this.isReady = false;
	}
}
