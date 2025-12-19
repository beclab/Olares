export class MessageQueue {
	private queue: Array<{
		id: string;
		tabId: string;
		type: string;
		payload: any;
	}> = [];
	private isProcessing = false;

	addMessage(message: {
		id: string;
		tabId: string;
		type: string;
		payload: any;
	}): void {
		this.queue.push(message);
	}

	async processQueue(
		handler: (message: {
			id: string;
			type: string;
			tabId: string;
			payload: any;
		}) => Promise<void>
	): Promise<void> {
		if (this.isProcessing) return;
		this.isProcessing = true;

		while (this.queue.length > 0) {
			const message = this.queue.shift();
			if (message) {
				try {
					await handler(message);
				} catch (error) {
					console.error(`Error processing message ${message.id}:`, error);
				}
			}
		}

		this.isProcessing = false;
	}
}
