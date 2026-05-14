import { App } from 'vue';

import BtPopover from './src/index.vue';

export { bus } from './src/popover';

export default (app: App) => {
	app.component(BtPopover.name!, BtPopover);
};
