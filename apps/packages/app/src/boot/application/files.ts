import { boot } from 'quasar/wrappers';
import { FilesApplication } from 'src/application/files';
import { setApplication } from '../../application/base';
export default boot(async () => {
	const files = new FilesApplication();
	setApplication(files);
});
