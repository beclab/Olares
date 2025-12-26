import { App } from 'vue';
import BtIcon from './BtIcon.vue';

const install = function (app: App<Element>): void {
	app.component('BtIcon', BtIcon);
};

export default install;
