<template>
	<div @click="selectFolder">
		<slot>
			<div
				class="transfer-to-select q-px-md row items-center justify-between text-ink-2"
				:style="{ '--select-height': selectHeight + 'px' }"
			>
				<div class="text-body3">
					{{ pathComputed }}
				</div>
				<q-icon name="sym_r_folder" size="16px" />
			</div>
		</slot>
	</div>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import DialogIndex from './../../../components/FilesDialog/DialogIndex.vue';
import { FilePath, PickType, useFilesStore } from '../../../stores/files';
import { computed, PropType, ref } from 'vue';
import { DriveType } from '../../../utils/interface/files';

const props = defineProps({
	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			return [
				DriveType.Drive,
				DriveType.Sync,
				DriveType.External,
				DriveType.Cache,
				DriveType.Data,
				DriveType.Common,
				DriveType.GoogleDrive
			];
		}
	},
	masterNode: {
		required: false,
		default: false,
		type: Boolean
	},
	selectHeight: {
		type: Number,
		required: false,
		default: 32
	},
	defaultPathNumber: {
		type: Number,
		required: false,
		default: 1
	}
});

const $q = useQuasar();

const filesStore = useFilesStore();

const selectFolder = () => {
	$q.dialog({
		component: DialogIndex,
		componentProps: {
			selectType: PickType.FOLDER,
			origins: props.origins,
			masterNode: props.masterNode
		}
	})
		.onOk(async (value: FilePath) => {
			filePath.value = value;
			emits('setSelectPath', value);
		})
		.onCancel(() => {
			emits('setSelectCancel');
		});
};

const filePath = ref<FilePath | undefined>(
	filesStore.currentPath[props.defaultPathNumber]
);

const pathComputed = computed(() => {
	if (filePath.value) {
		return (filePath.value as any).decodePath || filePath.value.path;
	}
	return '';
});

defineExpose({
	selectFolder
});

const emits = defineEmits(['setSelectPath', 'setSelectCancel']);
</script>

<style scoped lang="scss">
.transfer-to-select {
	height: var(--select-height);
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
