<template>
	<div class="full-width full-height">
		<bt-scroll-area
			class="full-width full-height"
			@scroll="onScroll"
			v-if="uploadList.length > 0"
		>
			<q-list>
				<transmission-entry-view
					v-for="(item, index) in uploadList"
					:key="item.id"
					:record="item"
					:selected="index === selectIndex"
					@on-selected-change="onSelectedChange(index)"
				/>
			</q-list>
			<footer-loading-component :has-data="loadMoreEnable" />
		</bt-scroll-area>
		<empty-view v-else class="source-root" />
	</div>
</template>

<script lang="ts" setup>
import TransmissionEntryView from '../../../../components/rss/entry/TransmissionEntryView.vue';
import FooterLoadingComponent from '../../../../components/files/FooterLoadingComponent.vue';
import EmptyView from '../../../../components/rss/EmptyView.vue';
import { useTransfer2Store } from '../../../../stores/transfer2';
import { computed, ref } from 'vue';

const selectIndex = ref(-1);
const showBarShadow = ref(false);
const loadMoreEnable = ref(false);
const transferStore = useTransfer2Store();

const onSelectedChange = async (index: number) => {
	selectIndex.value = index;
};

const onScroll = async (info: any) => {
	showBarShadow.value = info.verticalPosition > 0;
};

const uploadList = computed(() => {
	const list = transferStore.upload;
	if (list.length > 0) {
		return list
			.map((item) => {
				return transferStore.transferMap[item];
			})
			.sort((a, b) => b.startTime - a.startTime);
	} else {
		return [];
	}
});
</script>

<style scoped lang="scss"></style>
