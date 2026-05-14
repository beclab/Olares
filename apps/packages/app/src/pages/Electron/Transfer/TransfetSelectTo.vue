<template>
	<div @click="selectFolder">
		<slot>
			<div
				class="transfer-to-select q-px-md row items-center justify-between text-ink-2"
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
				DriveType.GoogleDrive
			];
		}
	},
	masterNode: {
		required: false,
		default: false,
		type: Boolean
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
			console.log(value);
			filePath.value = value;
			emits('setSelectPath', value);
		})
		.onCancel(() => {
			emits('setSelectCancel');
		});
};

const filePath = ref<FilePath | undefined>(filesStore.currentPath[1]);

const pathComputed = computed(() => {
	if (filePath.value) {
		return filePath.value.decodePath;
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
	height: 32px;
	border: 1px solid $separator;
	border-radius: 8px;
}
</style>
