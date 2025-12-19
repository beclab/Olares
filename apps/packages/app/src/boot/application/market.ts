import { boot } from 'quasar/wrappers';
import { setApplication } from '../../application/base';
import { MarketApplication } from '../../application/market';

export default boot(async () => {
	const market = new MarketApplication();
	setApplication(market);
});
