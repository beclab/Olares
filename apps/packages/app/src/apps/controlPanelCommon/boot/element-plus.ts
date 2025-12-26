import { boot } from 'quasar/wrappers';
import ElementUI from 'element-plus';
import 'element-plus/theme-chalk/index.css';

export default boot(({ app }) => {
	app.use(ElementUI);
});
