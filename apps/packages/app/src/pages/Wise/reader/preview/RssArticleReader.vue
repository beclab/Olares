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
import { Enclosure, FILE_TYPE } from '../../../../utils/rss-types';
import EnclosureCard from '../../../../components/rss/EnclosureCard.vue';
import FullContentReader from './FullContentReader.vue';
import RssAudioPlayer from './RssAudioPlayer.vue';
import RssVideoPlayer from './RssVideoPlayer.vue';
import { findEnclosure } from '../../../../api/wise';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { ref, watch } from 'vue';
import { useReaderStore } from '../../../../stores/rss-reader';

const transferStore = useTransferStore();
const readerStore = useReaderStore();
const enclosureList = ref<Enclosure[]>([]);

watch(
	readerStore.readingEntry,
	() => {
		if (readerStore.readingEntry) {
			findEnclosure(readerStore.readingEntry.id).then((list) => {
				enclosureList.value = list;
				transferStore.addEnclosureTasks(list);
			});
		}
	},
	{
		immediate: true
	}
);
</script>

<style scoped lang="scss"></style>
