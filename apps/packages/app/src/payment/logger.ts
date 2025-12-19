import { ref } from 'vue';
import type { LogItem, LogLevel } from './types';

export class Logger {
	private logs = ref<LogItem[]>([]);
	private maxLogCount = 1000;

	getLogs() {
		return this.logs;
	}

	private logBase(message: string, level: LogLevel): void {
		const timestamp = new Date().toLocaleTimeString();
		const logItem: LogItem = { time: timestamp, message, level };

		if (this.logs.value.length >= this.maxLogCount) {
			this.logs.value.shift();
		}

		this.logs.value.push(logItem);

		const consoleMethod = this.getConsoleMethod(level);
		consoleMethod(`[${timestamp}] [${level.toUpperCase()}] ${message}`);
	}

	private getConsoleMethod(level: LogLevel): (...args: any[]) => void {
		switch (level) {
			case 'warn':
				return console.warn;
			case 'error':
				return console.error;
			case 'success':
				return console.log;
			default:
				return console.log;
		}
	}

	log(message: string): void {
		this.logBase(message, 'info');
	}

	warn(message: string): void {
		this.logBase(message, 'warn');
	}

	error(message: string): void {
		this.logBase(message, 'error');
	}

	success(message: string): void {
		this.logBase(message, 'success');
	}

	clear(): void {
		this.logs.value = [];
	}
}

export const logger = new Logger();
