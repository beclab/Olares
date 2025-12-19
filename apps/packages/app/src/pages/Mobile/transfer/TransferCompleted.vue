<template>
	<q-list class="completed-list">
		<div v-for="history in historys" :key="history">
			<q-item class="history q-px-none" clickable @click="itemClick(history)">
				<q-item-section avatar style="min-width: 40px">
					<terminus-file-icon
						:name="transferStore.transferMap[history].name"
						:type="transferStore.transferMap[history].type"
						:path="transferStore.transferMap[history].path"
						:modified="false"
						:is-dir="transferStore.transferMap[history].isFolder"
					/>
				</q-item-section>

				<q-item-section>
					<q-item-label class="text-subtitle2 text-ink-1 content">{{
						transferStore.transferMap[history].name
					}}</q-item-label>
					<q-item-label class="text-body3 text-ink-3 content">
						<span>{{
							format.formatFileSize(transferStore.transferMap[history].size)
						}}</span>
						<span
							v-if="transferStore.transferMap[history].startTime"
							class="q-ml-sm"
							>{{
								formatDateFromNow(
									Number(transferStore.transferMap[history].startTime)
								)
							}}</span
						>
						<span
							class="q-ml-sm"
							v-if="
								transferStore.transferMap[history].front ===
								TransferFront.upload
							"
							>{{
								t('Upload to {address}', {
									address: transferStore.transferMap[history].path
								})
							}}</span
						>
					</q-item-label>
				</q-item-section>

				<q-item-section side @click.stop="deleteHistory(history)">
					<q-icon name="sym_r_close" size="20px" color="ink-2" />
				</q-item-section>
			</q-item>
			<!-- <q-separator color="separator" v-if="index + 1 < historys.length" /> -->
		</div>
	</q-list>
</template>

<script setup lang="ts">
import { PropType } from 'vue';
import { formatDateFromNow } from 'src/utils/format';
import { useTransfer2Store } from '../../../stores/transfer2';

import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import { useI18n } from 'vue-i18n';
import { format } from '../../../utils/format';
import { TransferFront } from '../../../utils/interface/transfer';

defineProps({
	historys: {
		type: Array as PropType<number[]>,
		require: true
	}
});

const transferStore = useTransfer2Store();

const deleteHistory = (id: number) => {
	transferStore.remove(id);
};

const { t } = useI18n();

const itemClick = (id: number) => {
	emits('itemClick', id);
};

const emits = defineEmits(['itemClick']);
</script>

<style scoped lang="scss">
.completed-list {
	.history {
		height: 72px;
	}
	.content {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
}
</style>
