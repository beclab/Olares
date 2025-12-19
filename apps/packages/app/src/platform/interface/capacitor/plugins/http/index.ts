import { registerPlugin } from '@capacitor/core';
import { CapacitorHttpPlugin } from '@capacitor/core/types/core-plugins';

const hookCapacitorHttp =
	registerPlugin<CapacitorHttpPlugin>('HookCapacitorHttp');

export { hookCapacitorHttp };
