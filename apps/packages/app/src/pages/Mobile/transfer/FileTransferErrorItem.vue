<template>
	<q-item class="history q-px-none" clickable @click="itemClick(file)">
		<q-item-section avatar style="min-width: 40px">
			<terminus-file-icon
				:name="transferStore.transferMap[file].name"
				:type="transferStore.transferMap[file].type"
				:path="transferStore.transferMap[file].path"
				:modified="false"
				:is-dir="transferStore.transferMap[file].isFolder"
			/>
		</q-item-section>

		<q-item-section>
			<q-item-label class="text-subtitle2 text-ink-1 content">{{
				transferStore.transferMap[file].name
			}}</q-item-label>
			<q-item-label class="text-body3 text-ink-3 content">
				<span>{{
					format.formatFileSize(transferStore.transferMap[file].size)
				}}</span>
			</q-item-label>
		</q-item-section>

		<q-item-section side @click.stop="deleteHistory(file)">
			<q-icon name="sym_r_close" size="20px" color="ink-2" />
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { useTransfer2Store } from '../../../stores/transfer2';
import TerminusFileIcon from '../../../components/common/TerminusFileIcon.vue';
import { format } from '../../../utils/format';

defineProps({
	file: {
		type: Number,
		required: true
	}
});

const transferStore = useTransfer2Store();

const deleteHistory = (id: number) => {
	transferStore.remove(id);
};

const itemClick = (id: number) => {
	emits('itemClick', id);
};

const emits = defineEmits(['itemClick']);
</script>

<style scoped lang="scss">
.processing-history {
	height: 72px;
	.name {
		text-overflow: ellipsis;
		white-space: nowrap;
		overflow: hidden;
	}
	.item-action {
		width: 32px;
		height: 32px;
	}
}
</style>
