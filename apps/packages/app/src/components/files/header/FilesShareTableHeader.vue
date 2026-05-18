<template>
	<div
		class="common-div"
		:style="origin_id === FilesIdType.PAGEID ? 'padding: 0 20px 0px 20px' : ''"
	>
		<div
			class="item header"
			:class="origin_id === FilesIdType.PAGEID ? '' : 'text-body3'"
			:style="{
				padding: origin_id === FilesIdType.PAGEID ? '20px 20px 3px 0' : '0 14px'
			}"
		>
			<div></div>
			<div>
				<p
					class="q-pl-xs"
					:class="{
						'share-name': true
					}"
					tabindex="0"
				>
					<span
						@click="sort(FilesSortType.NAME)"
						:title="$t('files.sortByName')"
						:aria-label="$t('files.sortByName')"
						>{{ $t('files.name') }}</span
					>
				</p>

				<p class="expiration-date" role="button" tabindex="0">
					<span>{{ $t('files.Expiration date') }}</span>
				</p>
				<p class="permission" role="button" tabindex="0">
					<span>{{ $t('files.permission') }}</span>
				</p>
				<p class="share-scope" role="button" tabindex="0">
					<span>{{ $t('files.Share scope') }}</span>
				</p>
				<p class="owner" role="button" tabindex="0">
					<span>{{ $t('files.Owner') }}</span>
				</p>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { getAppPlatform } from '../../../application/platform';
import { useFilesStore, FilesIdType } from './../../../stores/files';
import { FilesSortType } from './../../../utils/contact';
const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	}
	// byMe: {
	// 	type: Boolean,
	// 	required: true
	// }
});

const isPad = ref(getAppPlatform() && getAppPlatform().isPad);
const filesStore = useFilesStore();
const selectFileCount = ref(0);
if (isPad.value) {
	watch(
		() => filesStore.selected[props.origin_id],
		async () => {
			selectFileCount.value = filesStore.selected[props.origin_id].length;
		},
		{
			deep: true
		}
	);
}

const sort = async (by: FilesSortType) => {
	const current = filesStore.activeSort[props.origin_id];
	if (current.by == by) {
		filesStore.updateActiveSort(by, !current.asc, props.origin_id);
	} else {
		filesStore.updateActiveSort(by, true, props.origin_id);
	}
};
</script>

<style scoped lang="scss">
.common-div {
	width: 100%;
}
.mosaic-header {
	// width: 100%;
	height: 73px;
	border-bottom: 1px solid $separator;
}
.empty {
	img {
		width: 226px;
		height: 170px;
		margin-bottom: 20px;
	}
}
#listing {
	position: relative;
	overflow-y: scroll;
	&.hovered::-webkit-scrollbar-thumb {
		background-color: rgba(0, 0, 0, 0.2);
	}
}
.files-body {
	width: 100%;
	height: 100%;
}
.iconClockwise {
	animation: rotate 0.2s linear forwards;
}
@keyframes rotate {
	from {
		transform: rotate(0deg);
	}
	to {
		transform: rotate(180deg);
	}
}
.iconAnticlockwise {
	animation: reverse-rotate 0.2s linear forwards;
}
@keyframes reverse-rotate {
	0% {
		transform: rotate(180deg);
	}
	100% {
		transform: rotate(0deg);
	}
}
</style>
