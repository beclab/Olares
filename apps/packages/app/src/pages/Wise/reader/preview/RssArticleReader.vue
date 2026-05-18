<template>
	<div class="fill-height full-width" style="max-width: 100%">
		<div
			v-for="item in enclosureList"
			:key="item.id"
			:style="{
				'--padding': `${$q.screen.sm || $q.screen.xs ? 15 : 30}px`
			}"
		>
			<enclosure-card
				v-if="item.mime_type === FILE_TYPE.VIDEO"
				:enclosure="item"
			>
				<template v-slot:complete="{ path }">
					<rss-video-player :src="path" />
				</template>
			</enclosure-card>
			<enclosure-card
				v-if="item.mime_type === FILE_TYPE.AUDIO"
				:enclosure="item"
			>
				<template v-slot:complete="{ path }">
					<rss-audio-player :src="path" />
				</template>
			</enclosure-card>
		</div>
		<full-content-reader :margin-top="enclosureList.length > 0" />
	</div>
</template>

<script setup lang="ts">
import EnclosureCard from '../../../../components/rss/EnclosureCard.vue';
import FullContentReader from './FullContentReader.vue';
import RssAudioPlayer from './RssAudioPlayer.vue';
import RssVideoPlayer from './RssVideoPlayer.vue';
import { Enclosure, FILE_TYPE } from '../../../../utils/rss-types';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { useReaderStore } from '../../../../stores/rss-reader';
import { ref, watch, onMounted, onBeforeUnmount } from 'vue';
import { findEnclosure } from '../../../../api/wise';
import { busOff, busOn } from '../../../../utils/bus';

const transferStore = useTransferStore();
const readerStore = useReaderStore();
const enclosureList = ref<Enclosure[]>([]);

const loadEnclosures = () => {
	if (readerStore.readingEntry) {
		findEnclosure(readerStore.readingEntry.id).then((list) => {
			enclosureList.value = list;
			transferStore.addEnclosureTasks(list);
		});
	}
};

watch(
	() => readerStore.readingEntry?.id,
	() => {
		loadEnclosures();
	},
	{
		immediate: true
	}
);

const addEnclosure = (event: any) => {
	if (event.data && event.data?.entry_id === readerStore.readingEntry?.id) {
		const index = enclosureList.value.findIndex(
			(item) => item.id === event.data.id
		);
		if (index > -1) {
			enclosureList.value.splice(index, 1);
		} else {
			enclosureList.value.push(event.data);
		}
		transferStore.addEnclosureTasks([event.data]);
	}
};

// Reload enclosures when entry is retried/refreshed
onMounted(() => {
	busOn('entryRetry', loadEnclosures);
	busOn('enclosureUpdate', addEnclosure);
});

onBeforeUnmount(() => {
	busOff('entryRetry', loadEnclosures);
	busOff('enclosureUpdate', addEnclosure);
});
</script>

<style scoped lang="scss"></style>
