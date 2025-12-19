<template>
	<div
		v-if="restoring !== RestoreStatus.NONE"
		class="column justify-center items-center restore-root"
	>
		<div class="img-layout">
			<q-img class="app-logo" src="wise/icons/wise-128x128.png" />
			<div class="status-bg row justify-center items-center">
				<bt-loading
					v-if="restoring === RestoreStatus.LOADING"
					:loading="true"
					size="32px"
				/>
				<q-img
					v-if="restoring === RestoreStatus.SUCCESS"
					class="status-icon"
					:src="getRequireImage('wise/restore_success.svg')"
				/>

				<q-img
					v-if="restoring === RestoreStatus.FAIL"
					class="status-icon"
					:src="getRequireImage('wise/restore_fail.svg')"
				/>
			</div>
		</div>

		<div class="q-mt-xl text-ink-1 text-h5">{{ statusText }}</div>

		<q-btn
			v-if="
				restoring === RestoreStatus.FAIL || restoring === RestoreStatus.SUCCESS
			"
			class="close-btn text-ink-2"
			:label="t('close')"
			@click="onRestart"
			:loading="init"
		/>
	</div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { queryKnowledgeRestore } from 'src/api/wise/restore';
import { busOff, busOn } from 'src/utils/bus';
import BtLoading from '../base/BtLoading.vue';
import { getRequireImage } from 'src/utils/rss-utils';
import { useI18n } from 'vue-i18n';
import { date } from 'quasar';
import { sendMessageToWorker } from 'src/pages/Wise/database/sqliteService';
import { TERMINUS_ID } from 'src/utils/localStorageConstant';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

enum RestoreStatus {
	NONE = 'none',
	LOADING = 'running',
	SUCCESS = 'finish',
	FAIL = 'err'
}

const restoring = ref(RestoreStatus.NONE);
let restoreTimer: NodeJS.Timer | null = null;
let restoreOk = ref('');
const { t } = useI18n();
const init = ref();

const queryBackupStatus = () => {
	if (restoring.value !== RestoreStatus.NONE) {
		return;
	}
	restoring.value = RestoreStatus.LOADING;
	restoreTimer = setInterval(async () => {
		restoring.value = await queryKnowledgeRestore();
		if (
			(restoring.value === RestoreStatus.SUCCESS ||
				restoring.value === RestoreStatus.FAIL) &&
			!!restoreTimer
		) {
			localStorage.removeItem(TERMINUS_ID);
			restoreOk.value = date.formatDate(Date.now(), 'YYYY-MM-DD HH:mm');
			clearInterval(restoreTimer);
		}
	}, 10 * 1000);
};

onMounted(() => {
	busOn('appRestore', queryBackupStatus);
});

onBeforeUnmount(() => {
	if (restoreTimer) {
		clearInterval(restoreTimer);
	}
	busOff('appRestore', queryBackupStatus);
});

const statusText = computed(() => {
	switch (restoring.value) {
		case RestoreStatus.LOADING:
			return t('restoring_message');
		case RestoreStatus.SUCCESS:
			return t('restore_completed', { time: restoreOk.value });
		case RestoreStatus.FAIL:
			return t('restore_failed', { time: restoreOk.value });
		default:
			return '';
	}
});

const onRestart = () => {
	init.value = true;
	sendMessageToWorker('close')
		.then(() => {
			localStorage.removeItem(TERMINUS_ID);
			location.reload();
		})
		.catch((e: any) => {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: `Clear error, please refresh manually. Error: ${e?.message}`
			});
		})
		.finally(() => {
			init.value = false;
			restoring.value = RestoreStatus.NONE;
		});
};
</script>

<style scoped lang="scss">
.restore-root {
	z-index: 999;
	height: 100vh;
	width: 100vw;

	.img-layout {
		position: relative;

		.app-logo {
			width: 120px;
			height: 120px;
			border-radius: 27px;
		}

		.status-bg {
			background: $white;
			height: 40px;
			width: 40px;
			position: absolute;
			right: 0;
			bottom: 0;
			border-radius: 999px;

			.status-icon {
				width: 32px;
				height: 32px;
			}
		}
	}

	.close-btn {
		margin-top: 84px;
		height: 40px;
		border-radius: 8px;
		border: 1px solid $btn-stroke;
		min-width: 100px;
	}
}
</style>
