<template>
	<div class="files-body">
		<div
			class="empty column items-center justify-center full-height"
			v-if="
				filesStore.currentDirItems(origin_id)?.length +
					filesStore.currentFileItems(origin_id)?.length ==
				0
			"
		>
			<img src="./../../assets/nodata.svg" alt="empty" />
			<span
				class="text-body2 text-ink-1"
				v-if="filesStore.loading[origin_id]"
				>{{ $t('files.loading') }}</span
			>
			<span class="text-body2 text-ink-1" v-else>{{ $t('files.lonely') }}</span>
		</div>

		<div v-else id="listing" ref="listing" class="list file-icons">
			<BtScrollArea :style="`height: 100%`">
				<div class="common-div">
					<files-table-header :origin_id="origin_id" />

					<ListingItem
						v-for="item in filesStore.currentDirItems(origin_id)"
						:key="base64(item.name)"
						:item="item"
						viewMode="list"
						:selectType="selectType"
						:origin_id="origin_id"
					>
					</ListingItem>

					<ListingItem
						v-for="item in filesStore.currentFileItems(origin_id)"
						:key="base64(item.name)"
						:item="item"
						viewMode="list"
						:selectType="selectType"
						:origin_id="origin_id"
					>
					</ListingItem>
				</div>
			</BtScrollArea>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { useFilesStore, PickType } from './../../stores/files';
import ListingItem from '../../components/files/ListingItem.vue';
import FilesTableHeader from '../../components/files/FilesTableHeader.vue';
import { stringToBase64 } from '@didvault/sdk/src/core';

const filesStore = useFilesStore();

defineProps({
	origin_id: {
		type: Number,
		required: false
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: PickType.FOLDER
	}
});

const base64 = (name: string) => {
	return stringToBase64(name);
};
</script>

<style scoped lang="scss">
.empty {
	img {
		width: 226px;
		height: 170px;
		margin-bottom: 20px;
	}
}

.files-body {
	width: 100%;
	height: calc(100% - 40px);
}
</style>
