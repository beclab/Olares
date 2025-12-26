<template>
	<terminus-ebook v-if="path" :src="path" :played-time="playedTime" />
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import TerminusEbook from '../../../../components/common/TerminusEbook.vue';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { useReaderStore } from '../../../../stores/rss-reader';

const transferStore = useTransferStore();
const path = computed(() => {
	return transferStore.getDownloadUrl();
});

const playedTime = computed(() => {
	const readerStore = useReaderStore();
	if (readerStore.readingEntry) {
		return readerStore.readingEntry.played_time;
	} else {
		return 0;
	}
});
</script>
