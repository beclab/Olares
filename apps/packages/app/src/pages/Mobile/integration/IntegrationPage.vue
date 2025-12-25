<template>
	<div class="integration-root">
		<terminus-title-bar
			right-icon="sym_r_add"
			@on-right-click="addClick"
			:title="t('integration.title')"
		/>
		<terminus-scroll-area class="integration-scroll">
			<template v-slot:content>
				<q-list
					class="integration-list"
					v-if="integrationStore.accounts.length > 0"
				>
					<integration-item
						v-for="item in integrationStore.accounts"
						:key="`${item.type}_${item.name}`"
						:title="item.name"
						:available="item.available"
						:detail="`Authorized time:${formattedDate(item.create_at)}`"
						@account-click="clickCloud(item)"
					>
						<template v-slot:avatar>
							<q-img
								width="40px"
								height="40px"
								:noSpinner="true"
								:src="getAccountIcon(item)"
							/>
						</template>
					</integration-item>
				</q-list>
				<div class="integration-list column justify-center items-center" v-else>
					<img
						src="../../../assets/layout/nodata.svg"
						style="margin-top: 180px"
					/>
					<span class="q-mb-md text-ink-1" style="margin-top: 32px">
						{{ t('there_is_nothing_for_now') }}
					</span>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { useIntegrationStore } from '../../../stores/integration';
import { onMounted } from 'vue';
import IntegrationItem from './IntegrationItem.vue';
import { useI18n } from 'vue-i18n';

import integrationService from '../../../services/integration';
import { date } from 'quasar';
import { useRouter } from 'vue-router';
import { getRequireImage } from '../../../utils/imageUtils';
import { IntegrationAccountMiniData } from '@bytetrade/core';

const integrationStore = useIntegrationStore();

const { t } = useI18n();
const router = useRouter();

const addClick = () => {
	router.push('/integration/add');
};

function clickCloud(account: IntegrationAccountMiniData) {
	router.push(
		'/integration/common/detail/' + account.type + '/' + account.name
	);
}

onMounted(async () => {
	integrationStore.getAccount('all');
});

const getAccountIcon = (data: IntegrationAccountMiniData) => {
	const account = integrationService.getAccountByType(data.type);
	if (!account) {
		return '';
	}
	return getRequireImage(`setting/integration/${account.detail.icon}`);
};

const formattedDate = (datetime: number) => {
	if (datetime <= 0) {
		return '--';
	}
	return date.formatDate(datetime, 'YYYY-MM-DD HH:mm:ss');
};
</script>

<style scoped lang="scss">
.integration-root {
	width: 100%;
	height: 100%;

	.integration-scroll {
		width: 100%;
		height: calc(100% - 56px);
		.integration-list {
			padding-left: 20px;
			padding-right: 20px;
			width: 100%;

			.status-common {
				height: 20px;
				border-radius: 4px;
			}

			.status-available {
				background: $green-soft;
				color: $positive;
			}

			.status-unabled {
				background: $red-soft;
				color: $negative;
			}
		}
	}
}
</style>
