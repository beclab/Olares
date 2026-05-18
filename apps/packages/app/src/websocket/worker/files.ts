import { BaseWebsocketBean } from '../applications/base';
import { FilesWSType } from '../public/files';
export class FilesWebsocketBean extends BaseWebsocketBean {
	socketEntrance = false;

	lastCopyItemsData:
		| {
				type: FilesWSType;
				data: any;
		  }
		| undefined = undefined;

	otherTypeMethods(data: {
		type: FilesWSType;
		data: any;
		_port?: MessagePort;
	}): boolean {
		switch (data.type) {
			case FilesWSType.UpdateCopyItems:
			case FilesWSType.ResetCopyItems:
				{
					this.connections.forEach((port) =>
						port.postMessage({
							type: 'message',
							data: {
								type: data.type,
								data: data.data
							}
						})
					);
					this.lastCopyItemsData = {
						type: data.type,
						data: data.data
					};
				}
				return true;

			default:
				return false;
		}
	}
	addConnection(port: MessagePort): void {
		super.addConnection(port);
		if (this.lastCopyItemsData) {
			port.postMessage({
				type: 'message',
				data: this.lastCopyItemsData
			});
		}
	}
}
