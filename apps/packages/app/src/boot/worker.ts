import { boot } from 'quasar/wrappers';
import { workerInit } from '../pages/Wise/database/sqliteService';

export default boot(async () => {
	workerInit();
});
