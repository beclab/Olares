<template>
	<q-virtual-scroll
		ref="scroller"
		separator
		style="max-height: 100%; height: 100%"
		:items="array"
		v-show="array.length > 0"
		:virtual-scroll-item-size="132"
		:virtual-scroll-sticky-size-end="50"
		:virtual-scroll-slice-size="20"
		:virtual-scroll-slice-ratio-before="1.5"
		:virtual-scroll-slice-ratio-after="1.5"
	>
		<template v-slot:default="{ item, index }">
			<!--			<div>{{item}} {{index}}</div>-->
			<library-entry-view
				:entry="item"
				:key="item.id"
				:selected="getItemSelected(item.id, index)"
				:show-read-status="showReadStatus"
				:time-type="timeType"
				@on-entry-delete="onEntryDelete"
			/>
			<!--			<footer-loading-component-->
			<!--				v-if="array && index === array.length - 1"-->
			<!--				style="margin-bottom: 400px"-->
			<!--				:has-data="false"-->
			<!--			/>-->
		</template>
		<template v-slot:after>
			<footer-loading-component :has-data="false" />
		</template>
	</q-virtual-scroll>
	<!--	</bt-scroll-area>-->
	<empty-view class="list-view" v-show="array.length <= 0" />
</template>

<script lang="ts" setup>
import FooterLoadingComponent from '../../../../components/files/FooterLoadingComponent.vue';
import LibraryEntryView from '../../../../components/rss/entry/LibraryEntryView.vue';
import EmptyView from '../../../../components/rss/EmptyView.vue';
import { Entry } from '../../../../utils/rss-types';
import { computed, PropType, ref } from 'vue';
import { useRssStore } from '../../../../stores/rss';
import { useReaderStore } from '../../../../stores/rss-reader';

const readerStore = useReaderStore();
const rssStore = useRssStore();
const scroller = ref();

const props = defineProps({
	array: {
		type: Object as PropType<Entry[]>,
		default: () => [] as Entry[]
	},
	showReadStatus: {
		type: Boolean,
		default: true,
		require: true
	},
	timeType: {
		type: String,
		required: true
	}
});

const hasSelectedItem = computed(() => {
	if (readerStore.readingEntry && props.array && props.array.length > 0) {
		return (
			props.array.findIndex(
				(entry) => entry.id === readerStore.readingEntry.id
			) !== -1
		);
	}
	return false;
});

const getItemSelected = (id: string, index: number) => {
	if (hasSelectedItem.value) {
		return id === readerStore.readingEntry.id;
	}
	if (index === readerStore.readingIndex) {
		return true;
	}
	if (readerStore.readingIndex > props.array.length - 1) {
		return index === props.array.length - 1;
	}
};

const onEntryDelete = async (url: string, selected: boolean) => {
	await rssStore.removeEntry(url, selected);
};

// const onScroll = (info: any) => {
// 	showBarShadow.value = info.verticalPosition > 0;
// };
</script>

<style scoped lang="scss"></style>
