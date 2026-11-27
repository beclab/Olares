<template>
	<bt-list :label="index === 0 ? t('Applications') : ''">
		<div class="column q-px-lg q-py-md">
			<div class="overlay-app-item__top row justify-between items-center">
				<div class="row items-center">
					<q-img class="overlay-app-item__icon" no-spinner :src="app.icon" />
					<div class="column justify-start q-ml-sm">
						<div class="row justify-start items-center">
							<span
								class="text-ink-1"
								:class="{
									'text-body1': !deviceStore.isMobile,
									'text-subtitle3-m': deviceStore.isMobile
								}"
							>
								{{ app.title }}
							</span>
							<application-status
								class="q-ml-sm"
								ui-style="square"
								:status="app.status"
							/>
						</div>
					</div>
				</div>

				<div>
					<bt-loading v-if="isBusy" :loading="true" size="24px" />
					<bt-switch
						v-else
						truthy-track-color="blue-default"
						:model-value="app.overlay"
						:disable="overlayStore.isAppPending(app.name)"
						@update:model-value="emit('toggle', app.name, !app.overlay)"
					/>
				</div>
			</div>

			<div
				v-if="networkPortRows.length > 0"
				class="overlay-app-item__network column q-mt-md bg-background-6"
			>
				<div class="overlay-app-item__network-header row items-center">
					<q-icon name="sym_r_deployed_code" size="20px" color="ink-1" />
					<span class="text-subtitle2 text-ink-1 q-ml-xs">{{ app.title }}</span>
				</div>

				<div
					v-for="row in networkPortRows"
					:key="row.key"
					class="overlay-app-item__network-row row items-center justify-between"
				>
					<div class="row items-center no-wrap">
						<span class="text-body3 text-ink-2">{{ row.title }}</span>
						<q-icon
							v-if="row.description"
							name="sym_r_help"
							size="16px"
							color="ink-3"
							class="q-ml-xs"
						>
							<q-tooltip
								max-width="280px"
								anchor="bottom middle"
								self="top left"
							>
								<span class="text-body3 text-ink-tooltip">{{
									row.description
								}}</span>
							</q-tooltip>
						</q-icon>
					</div>
					<div
						class="row items-center no-wrap cursor-pointer q-ml-sm"
						@click="copyEndpoint(row.endpoint)"
					>
						<span class="text-body3 text-blue-default">{{ row.endpoint }}</span>
						<q-icon
							name="sym_r_content_copy"
							size="16px"
							color="blue-default"
							class="q-ml-xs"
						/>
					</div>
				</div>
			</div>
		</div>
	</bt-list>
</template>

<script lang="ts" setup>
import ApplicationStatus from 'src/components/settings/application/ApplicationStatus.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import BtLoading from 'src/components/base/BtLoading.vue';
import { getApplication } from 'src/application/base';
import { useDeviceStore } from 'src/stores/settings/device';
import { isDoingState } from 'src/constant/config';
import { notifySuccess } from 'src/utils/settings/btNotify';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';
import {
	OverlayApp,
	useOverlayGatewayStore
} from 'src/stores/settings/overlayGateway';

interface NetworkPortRow {
	key: string;
	title: string;
	description: string;
	endpoint: string;
}

const props = defineProps<{
	app: OverlayApp;
	index: number;
}>();

const { t } = useI18n();

const emit = defineEmits<{
	(e: 'toggle', name: string, enable: boolean): void;
}>();

const deviceStore = useDeviceStore();
const overlayStore = useOverlayGatewayStore();

const isBusy = computed(
	() =>
		isDoingState(props.app.status) || overlayStore.isAppPending(props.app.name)
);

const networkPortRows = computed((): NetworkPortRow[] => {
	if (!props.app.overlay) {
		return [];
	}

	const rows: NetworkPortRow[] = [];

	for (const network of props.app.underlay_networks) {
		for (const port of network.ports) {
			rows.push({
				key: `${network.ip}-${port.port}-${port.workload}`,
				title: port.title,
				description: port.description ?? '',
				endpoint: `${network.ip}:${port.port}`
			});
		}
	}

	return rows;
});

function copyEndpoint(endpoint: string) {
	getApplication()
		.copyToClipboard(endpoint)
		.then(() => {
			notifySuccess(t('copy_successfully'));
		});
}
</script>

<style scoped lang="scss">
.overlay-app-item {
	width: 100%;

	&__top {
		padding: 6px 0;
	}

	&__icon {
		width: 32px;
		height: 32px;
		border-radius: 8px;
	}

	&__network {
		padding: 12px 20px 20px;
		border-radius: 8px;
		gap: 20px;
	}

	&__network-header {
		padding-bottom: 12px;
		border-bottom: 1px solid $separator;
	}

	&__network-row {
		gap: 4px;
	}
}
</style>
