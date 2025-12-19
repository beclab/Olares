<template>
	<div class="file-footer" v-if="store.req">
		{{
			selectFileCount === 1
				? t('files.select_item', {
						count: selectFileCount || '',
						size: format.humanStorageSize(selectFileSize)
				  })
				: ''
		}}

		{{
			selectFileCount > 1
				? t('files.select_items', {
						count: selectFileCount || '',
						size: format.humanStorageSize(selectFileSize)
				  })
				: ''
		}}

		{{ totalFileCount }}
		{{ totalFileCount > 1 ? t('files.items') : t('files.item') }}
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { format } from '../../../utils/format';
import { useI18n } from 'vue-i18n';
import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const { t } = useI18n();
const store = useDataStore();
const filesStore = useFilesStore();

const selectFileCount = ref(0);
const selectFileSize = ref(0);
const totalFileCount = ref(0);

watch(
	() => filesStore.selected,
	async (newVaule) => {
		selectFileCount.value = newVaule.length;
		let totalSize = 0;
		if (newVaule.length && newVaule.length > 0) {
			for (let i = 0; i < newVaule.length; i++) {
				const el = newVaule[i];

				let filesSize: number =
					filesStore.currentFileList[props.origin_id]?.items[el].size || 0;
				totalSize += filesSize;
			}
		}

		selectFileSize.value = totalSize;
	},
	{
		deep: true
	}
);

watch(
	() => filesStore.currentFileList[props.origin_id],
	(newVal) => {
		if (newVal) {
			totalFileCount.value =
				newVal && newVal.items.length > 0 ? newVal.items.length : 0;
		}
	}
);
</script>

<style scoped lang="scss">
.file-footer {
	width: 100%;
	height: 40px;
	line-height: 40px;
	text-align: center;
	border-top: 1px solid $separator;
	justify-content: center !important;
	// border-left: solid 1px $separator;
}
</style>
