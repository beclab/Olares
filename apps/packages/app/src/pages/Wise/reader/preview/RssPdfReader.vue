<template>
	<div class="pdf-root column justify-center items-center">
		<pdfvuer
			:src="pdfConfigStore.source"
			v-for="i in pdfConfigStore.numPages"
			:key="i"
			:id="`pdf_${i}`"
			:page="i"
			:rotate="pdfConfigStore.rotate"
			v-model:scale="pdfConfigStore.scale"
			class="children"
			:annotation="true"
		>
			<template v-slot:loading>
				<div class="loading column justify-center items-center">
					<bt-loading :loading="true" size="76px" />
				</div>
			</template>
		</pdfvuer>
	</div>
</template>

<script lang="ts" setup>
import { useReadingProgressStore } from '../../../../stores/rss-reading-progress';
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import BtLoading from '../../../../components/base/BtLoading.vue';
import { useTransferStore } from '../../../../stores/rss-transfer';
import { useReaderStore } from '../../../../stores/rss-reader';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { usePDfStore } from '../../../../stores/pdf';
import { bus } from '../../../../utils/bus';
import pdfvuer from 'pdfvuer';

const readingProgressStore = useReadingProgressStore();
const transferStore = useTransferStore();
const readerStore = useReaderStore();
const pdfConfigStore = usePDfStore();
let initiated = ref(false);

watch(
	readerStore.readingEntry,
	() => {
		updatePdfData();
	},
	{
		immediate: true
	}
);

function scrollIntoPos() {
	const position = document.getElementById(`pdf_${pdfConfigStore.pageNum}`);
	if (position) {
		position?.scrollIntoView();
	}
}

const playedTime = computed(() => {
	const readerStore = useReaderStore();
	if (readerStore.readingEntry) {
		return readerStore.readingEntry.played_time;
	} else {
		return 0;
	}
});

onMounted(() => {
	bus.on('scrollIntoPos', scrollIntoPos);
	updatePdfData();
});

onBeforeUnmount(() => {
	bus.off('scrollIntoPos', scrollIntoPos);
});

function updatePdfData() {
	if (readerStore.readingEntry) {
		pdfConfigStore.topicLoad = false;
		pdfConfigStore.source = transferStore.getDownloadUrl();
		pdfvuer
			.createLoadingTask(pdfConfigStore.source)
			.then((pdf) => {
				initiated.value = true;
				pdfConfigStore.numPages = pdf.numPages;
				readingProgressStore.setTotalProgress(pdf.numPages);
				if (playedTime.value) {
					pdfConfigStore.skipPage(playedTime.value);
				}
				setTimeout(() => {
					pdfConfigStore.topicLoad = true;
				}, 500);
			})
			.catch((e: any) => {
				console.log(e);
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: e ? e.message : 'Load PDF Error'
				});
			});
	} else {
		pdfConfigStore.init();
	}
}

function findPos(obj) {
	return obj.offsetTop;
}

function changePageByScroll(scrollY: number) {
	if (!initiated.value) {
		return;
	}
	let i = 1,
		count = Number(pdfConfigStore.numPages),
		pageNum = -1;
	do {
		if (
			i < count &&
			scrollY >= findPos(document.getElementById(`pdf_${i}`)) &&
			scrollY <= findPos(document.getElementById(`pdf_${i + 1}`))
		) {
			pageNum = i;
			break;
		}
		if (
			i === count &&
			scrollY >= findPos(document.getElementById(`pdf_${i}`))
		) {
			pageNum = i;
			break;
		}
		i++;
	} while (i <= count);
	if (pageNum !== -1) {
		pdfConfigStore.pageNum = pageNum;
		readingProgressStore.updateProgress(pageNum);
	}
}

defineExpose({
	changePageByScroll
});
</script>

<style scoped lang="scss">
.pdf-root {
	overflow: scroll;
	height: 100%;
	width: 100%;
	padding-bottom: 10px;

	.children {
		width: auto;
		box-shadow: 0 4px 10px 0 #0000001a;
		border-radius: 12px;
		overflow: hidden;
		border: 1px solid $separator;
		margin-bottom: 20px;
	}

	.loading {
		width: 720px;
		height: 869px;
	}
}
</style>
