<template>
	<div class="integration-add-root">
		<terminus-title-bar :title="t('integration.add_account')" />
		<terminus-scroll-area class="integration-add-scroll">
			<template v-slot:content>
				<q-list class="integration-list">
					<terminus-item
						v-for="item in items"
						:key="`${item.type}`"
						:show-board="true"
						:image-path="getAccountIcon(item)"
						:item-height="64"
						:icon-size="40"
						class="q-mt-md"
						@click="accountCreate(item)"
					>
						<template v-slot:title>
							{{ item.detail.name }}
						</template>
						<template v-slot:side>
							<q-icon
								name="sym_r_keyboard_arrow_right"
								size="20px"
								color="ink-3"
							/>
						</template>
					</terminus-item>
				</q-list>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import integrationService from '../../../services/integration/index';
import {
	AccountAddMode,
	IntegrationAccountInfo
} from '../../../services/abstractions/integration/integrationService';
import TerminusItem from '../../../components/common/TerminusItem.vue';
import { useRoute, useRouter } from 'vue-router';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useQuasar } from 'quasar';

const route = useRoute();
const { t } = useI18n();
const isBackup = Number(route.query.backup) == 1;

console.log(route);

const items = ref(
	isBackup
		? integrationService.supportBackupList
		: integrationService.supportAuthList
);
const router = useRouter();
const $q = useQuasar();

const getAccountIcon = (account: IntegrationAccountInfo) => {
	return `setting/integration/${account.detail.icon}`;
};

const accountCreate = async (item: IntegrationAccountInfo) => {
	$q.loading.show();
	const webSupport = await integrationService.requestIntegrationAuth(
		item.type,
		{
			router,
			backup: isBackup
		}
	);
	$q.loading.hide();
	if (webSupport.addMode == AccountAddMode.common) {
		if (webSupport.status) {
			router.back();
		} else if (webSupport.message) {
			notifyFailed(webSupport.message);
		}
	}
};
</script>

<style scoped lang="scss">
.integration-add-root {
	width: 100%;
	height: 100%;
	.integration-add-scroll {
		height: calc(100% - 56px);
		width: 100%;
		.integration-list {
			padding-left: 20px;
			padding-right: 20px;
			width: 100%;
			// height: 100%;
		}
	}
}
</style>
