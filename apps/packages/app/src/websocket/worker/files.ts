import { BaseWebsocketBean } from '../applications/base';
import { FilesWSType } from '../public/files';

export class FilesWebsocketBean extends BaseWebsocketBean {
	socketEntrance = false;

	otherTypeMethods(data: {
		type: FilesWSType;
		data: any;
		_port?: MessagePort;
	}): boolean {
		switch (data.type) {
			case FilesWSType.UpdateCopyItems:
			case FilesWSType.ResetCopyItems:
				// Copy/cut clipboard is durable cross-tab state: retain it so a
				// newly opened Files tab receives the current selection on join.
				this.publishAppMessage(
					{
						type: data.type,
						data: data.data
					},
					true
				);
				return true;

			default:
				return false;
		}
	}
}
