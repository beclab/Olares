import { QVueGlobals } from 'quasar';
import {
	AccountAddMode,
	TencentIntegrationAccount,
	OperateIntegrationAuth
} from '../abstractions/integration/integrationService';

import { AccountType, IntegrationAccountMiniData } from '@bytetrade/core';
import { useDeviceStore } from 'src/stores/settings/device';
import AddTencentDialog from '../../pages/settings/Integration/dialog/AddAccountDialog.vue';
import { useIntegrationStore } from 'src/stores/settings/integration';

export class TencentAuthService extends OperateIntegrationAuth<TencentIntegrationAccount> {
	type = AccountType.Tencent;
	addMode = AccountAddMode.direct;

	async signIn(options: any) {
		const quasar = options.quasar as QVueGlobals;
		const router = options.router;
		const deviceStore = useDeviceStore();
		if (deviceStore.isMobile) {
			router.push({
				path: '/integration/account/add',
				query: {
					accountType: this.type,
					backup: options.backup
				}
			});
		} else {
			quasar
				.dialog({
					component: AddTencentDialog,
					componentProps: {
						accountType: this.type,
						backup: options.backup
					}
				})
				.onOk(() => {
					const integrationStore = useIntegrationStore();
					integrationStore.getAccount('all');
				});
		}
		return undefined;
	}

	async permissions() {
		return {
			title: '',
			scopes: []
		};
	}

	async webSupport() {
		return {
			status: true,
			message: ''
		};
	}

	detailPath(account: IntegrationAccountMiniData) {
		return '/integration/common/detail/' + account.type + '/' + account.name;
	}
}
