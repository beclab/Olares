<template>
	<div class="pdf-root column justify-center items-center">
		<div
			v-if="!initiated"
			class="initial-loading column justify-center items-center"
		>
			<div class="loading-card column justify-center items-center">
				<bt-loading :loading="true" size="64px" />
				<div class="loading-text">{{ loadingStatusText }}</div>
				<div v-if="loadingSubText" class="loading-sub-text">
					{{ loadingSubText }}
				</div>
				<div v-if="loadingProgress > 0" class="loading-progress">
					<div class="progress-bar" :style="{ width: loadingProgress + '%' }" />
				</div>
			</div>
		</div>

		<template v-else v-for="i in pageNumbers" :key="i">
			<pdfvuer
				v-if="shouldRenderPage(i)"
				:src="pdfConfigStore.source"
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
			<div v-else :id="`pdf_${i}`" class="children placeholder">
				<div class="loading column justify-center items-center">
					<span class="page-number">{{ i }}</span>
				</div>
			</div>
		</template>
	</div>
</template>

<script lang="ts" setup>
import BtLoading from '../../../../../components/base/BtLoading.vue';
import { useReadingProgressStore } from 'src/stores/rss-reading-progress';
import { configurePdfWorker } from 'src/pages/Wise/reader/pdf';
import { clearThumbnailCache } from '../services/pdfThumbnail';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useTransferStore } from 'src/stores/rss-transfer';
import { useReaderStore } from 'src/stores/rss-reader';
import { bus } from 'src/utils/bus';
import { useI18n } from 'vue-i18n';
import pdfvuer from 'pdfvuer';
import {
	PDF_LAZY_LOAD_PRESETS,
	usePdfLazyLoad
} from 'src/pages/Wise/reader/pdf';
import {
	fetchPDFWithProgress,
	pdfCache,
	formatBytes,
	formatSpeed,
	DownloadProgress
} from '../services/pdfCache';
import {
	computed,
	nextTick,
	onBeforeUnmount,
	onMounted,
	ref,
	watch
} from 'vue';

configurePdfWorker();

const readingProgressStore = useReadingProgressStore();
const transferStore = useTransferStore();
const readerStore = useReaderStore();
let initiated = ref(false);

const isLoading = ref(false);
const currentLoadingUrl = ref('');
const loadId = ref(0);

type LoadingStage = 'idle' | 'checking' | 'fetching' | 'parsing' | 'rendering';
const loadingStage = ref<LoadingStage>('idle');
const loadingProgress = ref(0);
const isFromCache = ref(false);
const downloadProgress = ref<DownloadProgress | null>(null);
const { t } = useI18n();

// Main loading status text
const loadingStatusText = computed(() => {
	switch (loadingStage.value) {
		case 'checking':
			return t('Checking cache...');
		case 'fetching':
			if (isFromCache.value) {
				return t('Reading from cache...');
			}
			// Show download progress details
			if (downloadProgress.value) {
				const { loaded, total, speed } = downloadProgress.value;
				if (total > 0) {
					const percent = Math.min(99, Math.round((loaded / total) * 100));
					return `${percent}% (${formatBytes(loaded)} / ${formatBytes(
						total
					)}) ${formatSpeed(speed)}`;
				}
				return `${formatBytes(loaded)} ${formatSpeed(speed)}`;
			}
			return t('Downloading document...');
		case 'parsing':
			return t('Parsing document...');
		case 'rendering':
			return t('Preparing pages...');
		default:
			return t('Loading...');
	}
});

// Sub-text for additional context
const loadingSubText = computed(() => {
	if (loadingStage.value === 'fetching' && !isFromCache.value) {
		return t('First time loading, will be cached for instant access');
	}
	return '';
});

const {
	visiblePages,
	pageNumbers,
	updateVisiblePages,
	shouldRenderPage,
	pdfConfigStore
} = usePdfLazyLoad(PDF_LAZY_LOAD_PRESETS.reader);

watch(
	() => readerStore.readingEntry,
	(newEntry, oldEntry) => {
		if (newEntry?.id !== oldEntry?.id) {
			console.log('PDF: readingEntry change，reload');
			updatePdfData();
		}
	}
);

function scrollIntoPos() {
	const targetPage = pdfConfigStore.pageNum;
	console.log('PDF: scrollIntoPos call，targetPage:', targetPage);
	const position = document.getElementById(`pdf_${targetPage}`);
	if (position) {
		console.log('PDF: find pdf element by id，scrollIntoView');
		position.scrollIntoView({ behavior: 'auto', block: 'start' });
	} else {
		console.warn('PDF: don‘t find pdf_' + targetPage);
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
	console.log('PDF: destroy，clear resource...');
	bus.off('scrollIntoPos', scrollIntoPos);

	if (pdfConfigStore.pageNum > 0) {
		console.log('PDF: updateProgress，pageNum:', pdfConfigStore.pageNum);
		readingProgressStore.updateProgress(pdfConfigStore.pageNum);
	}

	loadId.value++;

	destroyPdfDocument();

	clearThumbnailCache();

	initiated.value = false;
	isLoading.value = false;
	currentLoadingUrl.value = '';
	visiblePages.value = [];

	pdfConfigStore.source = null;
	pdfConfigStore.numPages = 0;
	pdfConfigStore.init();

	console.log('PDF: clear resource ok');
});

function destroyPdfDocument() {
	if (pdfConfigStore.pdfDocument) {
		try {
			console.log('PDF: destroy Pdf...');
			pdfConfigStore.pdfDocument.destroy();
			pdfConfigStore.pdfDocument = null;
		} catch (e) {
			console.warn('PDF: destroy Pdf failed', e);
		}
	}
}

async function updatePdfData() {
	if (!readerStore.readingEntry) {
		pdfConfigStore.init();
		return;
	}

	const pdfUrl = transferStore.getDownloadUrl();

	if (isLoading.value && currentLoadingUrl.value === pdfUrl) {
		console.log('PDF: already loading, skipping duplicate request');
		return;
	}

	if (initiated.value && currentLoadingUrl.value === pdfUrl) {
		console.log('PDF: already loaded, skipping duplicate request');
		return;
	}

	const thisLoadId = ++loadId.value;
	isLoading.value = true;
	currentLoadingUrl.value = pdfUrl;
	pdfConfigStore.topicLoad = false;

	loadingStage.value = 'checking';
	loadingProgress.value = 5;
	isFromCache.value = false;

	const timeStart = performance.now();
	console.log('PDF: ========== Start loading (ID:', thisLoadId, ') ==========');

	try {
		isFromCache.value = await pdfCache.has(pdfUrl);

		loadingStage.value = 'fetching';
		loadingProgress.value = 10;
		downloadProgress.value = null;

		const t1 = performance.now();
		const { data: pdfData, fromCache } = await fetchPDFWithProgress(pdfUrl, {
			onProgress: (progress) => {
				downloadProgress.value = progress;
				// Map download progress to 10-40% of loading bar
				loadingProgress.value = 10 + Math.round(progress.percent * 0.3);
			}
		});

		if (loadId.value !== thisLoadId) {
			console.log('PDF: Load cancelled (ID:', thisLoadId, ')');
			return;
		}

		loadingProgress.value = 40;
		const t2 = performance.now();
		console.log(
			`PDF: [Step 1] Fetched data: ${(t2 - t1).toFixed(0)}ms, size: ${(
				pdfData.byteLength /
				1024 /
				1024
			).toFixed(2)} MB, fromCache: ${fromCache}`
		);

		loadingStage.value = 'parsing';
		loadingProgress.value = 50;

		destroyPdfDocument();

		const t3 = performance.now();
		// Mitigates Dependabot advisory #142 (GHSA on pdfjs-dist <= 4.1.392 reachable
		// via pdfvuer@2.0.1's hard-pin of pdfjs-dist@2.5.207). Disables eval-based
		// PostScript function compilation in pdf.js so a malicious PDF cannot execute
		// attacker-controlled JavaScript in the host origin. pdfvuer's
		// createLoadingTask forwards source fields to pdf.js getDocument(), which
		// honors this flag; both load paths (this explicit call and the <pdfvuer>
		// component's :src binding) share this same source object.
		pdfConfigStore.source = {
			data: new Uint8Array(pdfData),
			isEvalSupported: false
		};
		const t4 = performance.now();
		console.log(`PDF: [Step 2] Set source: ${(t4 - t3).toFixed(0)}ms`);

		loadingProgress.value = 60;
		const t5 = performance.now();
		console.log('PDF: [Step 3] Parsing PDF...');
		const pdf = await pdfvuer.createLoadingTask(pdfConfigStore.source);

		pdfConfigStore.pdfDocument = pdf;

		if (loadId.value !== thisLoadId) {
			console.log('PDF: Load cancelled (ID:', thisLoadId, ')');
			if (pdf && pdf.destroy) {
				pdf.destroy();
				pdfConfigStore.pdfDocument = null;
			}
			return;
		}

		loadingProgress.value = 80;
		const t6 = performance.now();
		console.log(
			`PDF: [Step 3] PDF parsed: ${(t6 - t5).toFixed(0)}ms, ${
				pdf.numPages
			} pages total`
		);

		await pdfConfigStore.extractOutline(pdf);

		loadingStage.value = 'rendering';
		loadingProgress.value = 90;

		const t7 = performance.now();
		pdfConfigStore.numPages = pdf.numPages;
		readingProgressStore.setTotalProgress(pdf.numPages);

		if (playedTime.value && playedTime.value > 0) {
			const targetPage = Math.min(playedTime.value, pdf.numPages);
			console.log(
				'PDF: Restoring reading progress, jumping to page:',
				targetPage
			);
			pdfConfigStore.pageNum = targetPage;
		} else if (!pdfConfigStore.pageNum || pdfConfigStore.pageNum < 1) {
			pdfConfigStore.pageNum = 1;
		}

		readingProgressStore.updateProgress(pdfConfigStore.pageNum);

		updateVisiblePages();

		initiated.value = true;

		console.log('PDF: Scrolling to page:', pdfConfigStore.pageNum);
		nextTick(() => {
			setTimeout(() => {
				scrollIntoPos();
			}, 300);
		});
		const t8 = performance.now();
		console.log(`PDF: [Step 4] State updated: ${(t8 - t7).toFixed(0)}ms`);

		loadingProgress.value = 100;

		setTimeout(() => {
			pdfConfigStore.topicLoad = true;
		}, 500);

		const timeEnd = performance.now();
		console.log(
			`PDF: ========== Loaded, total elapsed: ${(timeEnd - timeStart).toFixed(
				0
			)}ms ==========`
		);
	} catch (e: any) {
		if (loadId.value !== thisLoadId) {
			return;
		}
		console.log(e);
		BtNotify.show({
			type: NotifyDefinedType.FAILED,
			message: e ? e.message : 'Load PDF Error'
		});
	} finally {
		if (loadId.value === thisLoadId) {
			isLoading.value = false;
			loadingStage.value = 'idle';
		}
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
		const currentPos = findPos(document.getElementById(`pdf_${i}`));
		const nextPos =
			i < count ? findPos(document.getElementById(`pdf_${i + 1}`)) : Infinity;

		if (scrollY >= currentPos && scrollY < nextPos) {
			pageNum = i;
			break;
		}
		i++;
	} while (i <= count);

	if (pageNum !== -1 && pageNum !== pdfConfigStore.pageNum) {
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

	.initial-loading {
		width: 100%;
		height: 100%;
		min-height: 400px;

		.loading-card {
			background: rgba(255, 255, 255, 0.95);
			border-radius: 16px;
			padding: 40px 60px;
			box-shadow: 0 8px 32px rgba(0, 0, 0, 0.08);
			backdrop-filter: blur(10px);
			gap: 20px;
			min-width: 280px;
		}

		.loading-text {
			color: #666;
			font-size: 15px;
			font-weight: 500;
			margin-top: 8px;
		}

		.loading-sub-text {
			color: #999;
			font-size: 12px;
			margin-top: 4px;
			text-align: center;
			max-width: 250px;
		}

		.loading-progress {
			width: 200px;
			height: 4px;
			background: #e8e8e8;
			border-radius: 2px;
			overflow: hidden;
			margin-top: 8px;

			.progress-bar {
				height: 100%;
				background: linear-gradient(90deg, #ff9500, #ff6b00);
				border-radius: 2px;
				transition: width 0.3s ease;
			}
		}
	}

	.children {
		width: auto;
		box-shadow: 0 4px 10px 0 #0000001a;
		border-radius: 12px;
		overflow: hidden;
		border: 1px solid $separator;
		margin-bottom: 20px;
	}

	.placeholder {
		background: linear-gradient(135deg, #f5f7fa 0%, #e4e8ec 100%);
		display: flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		max-width: 800px;
		height: 1000px;
	}

	.page-number {
		color: #bbb;
		font-size: 24px;
		font-weight: 500;
	}

	.loading {
		width: 720px;
		height: 869px;
	}
}
</style>
