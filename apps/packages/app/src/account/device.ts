import { Serializable } from './serialization';

export class DeviceInfo extends Serializable {
	get kind() {
		return 'deviceinfo';
	}

	/** Platform/Operating System running on the device */
	platform = '';

	/** OS version running on the device */
	osVersion = '';

	/** Unique device identifier */
	id = '';

	/** Padloc version installed on the device */
	appVersion = '';

	vendorVersion = '';

	/** The user agent of the browser running the application */
	userAgent = '';

	/** The devices locale setting */
	locale = 'en';

	/** The device manufacturer, if available */
	manufacturer = '';

	/** The device mode, if available */
	model = '';

	/** The browser the application was loaded in, if applicable */
	browser = '';

	/** The version of the browser the application was loaded in, if applicable */
	browserVersion = '';

	description = '';

	runtime = '';

	constructor(props?: Partial<DeviceInfo>) {
		super();
		props && Object.assign(this, props);
	}
}
