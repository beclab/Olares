<template>
	<q-btn
		@click.stop
		:color="theme?.btnMoreDefaultColor"
		padding="6px"
		outline
		class="download-action-btn"
		:text-color="theme?.btnMoreTextDefaultColor"
		v-if="download_status !== DownloadStatusEnum.COMPLETE"
	>
		<q-icon
			name="sym_r_more_horiz"
			:color="theme?.btnMoreTextDefaultColor"
			size="20px"
		/>
		<bt-tooltip
			anchor="top middle"
			self="bottom middle"
			:label="$t('base.more')"
		/>
		<bt-popup style="width: 176px">
			<bt-popup-item
				v-if="download_status === DownloadStatusEnum.PAUSED"
				v-close-popup
				icon="sym_r_download"
				:title="$t('download.resume')"
				@on-item-click="retryHandler"
			/>
			<bt-popup-item
				v-if="
					download_status === DownloadStatusEnum.CANCEL ||
					download_status === DownloadStatusEnum.ERROR
				"
				v-close-popup
				icon="sym_r_autorenew"
				:title="$t('download.retry')"
				@on-item-click="retryHandler"
			/>
			<bt-popup-item
				v-if="
					download_status === DownloadStatusEnum.DOWNLOADING ||
					download_status === DownloadStatusEnum.WAITING
				"
				v-close-popup
				icon="sym_r_pause"
				:title="$t('download.pause')"
				@on-item-click="pauseHandler"
			/>
			<bt-popup-item
				v-if="
					download_status === DownloadStatusEnum.DOWNLOADING ||
					download_status === DownloadStatusEnum.WAITING ||
					download_status === DownloadStatusEnum.PAUSED
				"
				v-close-popup
				icon="sym_r_block"
				:title="$t('base.cancel')"
				@on-item-click="cancelHandler"
			/>
			<bt-popup-item
				v-if="
					download_status === DownloadStatusEnum.DOWNLOADING ||
					download_status === DownloadStatusEnum.PAUSED ||
					download_status === DownloadStatusEnum.ERROR ||
					download_status === DownloadStatusEnum.WAITING ||
					download_status === DownloadStatusEnum.CANCEL
				"
				v-close-popup
				icon="sym_r_open_in_new"
				:title="$t('base.open_in_wise')"
				@on-item-click="openFile"
			/>
		</bt-popup>
	</q-btn>

	<!-- <q-btn
		class="btn-size-sm btn-no-text btn-no-border btn-circle-border"
		color="ink-2"
		outline
		no-caps
		@click.stop
		@click="onRemove"
		icon="sym_r_do_not_disturb_on"
	>
		<bt-tooltip :label="$t('base.remove')" />
	</q-btn> -->
</template>

<script setup lang="ts">
import BtPopup from 'src/components/base/BtPopup.vue';
import { DownloadStatusEnum } from '../../types/commonApi';
import BtPopupItem from 'src/components/base/BtPopupItem.vue';
import BtTooltip from 'src/components/base/BtTooltip.vue';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';

import { DOWNLOAD_OPERATE } from 'src/utils/rss-types';
import { downloadTaskOperate } from 'src/api/wise/download';
import { inject } from 'vue';
import { COLLECT_THEME } from 'src/constant/provide';
import { getApplication } from 'src/application/base';
import { useUserStore } from 'src/stores/user';
import { replaceOriginDomain } from 'src/utils/url2';
import { openUrl } from 'src/utils/bex/tabs';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);

const collectSiteStore = useCollectSiteStore();
const appAbilitiesStore = useAppAbilitiesStore();

interface Props {
	download_status: `${DownloadStatusEnum}`;
	task_id: number | string;
}

const props = withDefaults(defineProps<Props>(), {});

const pauseHandler = async () => {
	try {
		await downloadTaskOperate(props.task_id.toString(), DOWNLOAD_OPERATE.PAUSE);

		collectSiteStore.updateDownloadStatusByTaskId(
			Number(props.task_id),
			DownloadStatusEnum.PAUSED
		);
	} catch (e) {
		//
	}
};

const cancelHandler = async () => {
	try {
		await downloadTaskOperate(
			props.task_id.toString(),
			DOWNLOAD_OPERATE.CANCEL
		);
		collectSiteStore.updateDownloadStatusByTaskId(
			Number(props.task_id),
			DownloadStatusEnum.CANCEL
		);
	} catch (e) {
		//
	}
};

const retryHandler = async () => {
	try {
		await downloadTaskOperate(props.task_id.toString(), DOWNLOAD_OPERATE.RETRY);
		collectSiteStore.updateDownloadStatusByTaskId(
			Number(props.task_id),
			DownloadStatusEnum.WAITING
		);
	} catch (e) {
		//
	}
};

const openFile = (file = '') => {
	let url = '';
	const filesKey = 'wise';
	if (getApplication().platform && getApplication().platform?.isClient) {
		url = appAbilitiesStore.getAppDomain(filesKey, '/download');
	} else {
		const origin = document.location.origin;
		url = `${origin}/download`;
	}

	openUrl(url);
};
</script>

<style lang="scss">
.download-action-btn.q-btn--outline::before {
	border-color: $btn-stroke;
}
</style>
