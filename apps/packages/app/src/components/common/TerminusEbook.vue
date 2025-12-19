<template>
	<!-- <div class="toc full-width">
		<li
			v-for="chapter in bookToc"
			:key="chapter.id"
			@click="goToChapter(chapter)"
		>
			{{ chapter.label }}
		</li>
	</div> -->
	<div id="viewer" ref="viewerRef" />

	<!--	<button @click="prevPage">Previous Page</button>-->
	<!--	<button @click="nextPage">Next Page</button>-->
	<!--	<input type="text" v-model="pageNumber" placeholder="Enter page number" />-->
	<!--	<button @click="gotoPage">Go to Page</button>-->
</template>

<script lang="ts" setup>
import { onMounted, ref, onBeforeUnmount } from 'vue';
import { useReadingProgressStore } from '../../stores/rss-reading-progress';
const book = ref();
const rendition = ref();
const viewerRef = ref();
const selectedChapter = ref();
const readingProgressStore = useReadingProgressStore();

const props = defineProps({
	src: {
		type: String,
		required: true
	},
	playedTime: {
		string: String,
		required: false
	}
});

async function loadBook() {
	const ePub = (await import('epubjs')).default;
	book.value = ePub(props.src);
	rendition.value = book.value.renderTo(viewerRef.value, {
		manager: 'continuous',
		flow: 'scrolled',
		width: '100%',
		height: '100%',
		allowScriptedContent: true
	});

	const displayed = rendition.value.display(selectedChapter.value);

	displayed.then(function () {
		// -- do stuff
	});

	book.value.ready
		.then(() => {
			return book.value.locations.generate();
		})
		.then(async () => {
			const totalLocations = book.value.locations.length();
			console.log(totalLocations);
			readingProgressStore.setTotalProgress(1, totalLocations);
			if (props.playedTime) {
				console.log('jump', props.playedTime);
				rendition.value.display(props.playedTime);
			}
		})
		.then(() => {
			rendition.value.on('relocated', (location) => {
				const currentLocation = book.value.locations.percentageFromCfi(
					location.start.cfi
				);
				console.log(currentLocation);
				readingProgressStore.updateProgress(currentLocation);
			});
		});

	// Navigation loaded
	book.value.loaded.navigation.then(function (toc) {
		console.log(toc);
	});
}

// async function getChapterText(book, cfiRange) {
// 	const range = await book.value.getRange(cfiRange);
// 	return range ? range.toString() : '';
// }

// function prevPage() {
// 	rendition.prev();
// }
// function nextPage() {
// 	rendition.next();
// }
// function gotoPage() {
// 	rendition.display(pageNumber.value);
// }

// function goToChapter(chapter) {
// 	let id = chapter.id;
// 	if (id.includes('#')) {
// 		const array = chapter.id.split('#');
// 		id = array[1];
// 	}
// 	const contents = rendition.value.getContents();
// 	for (let i = 0; i < contents.length; i++) {
// 		const iframe = contents[i].document.defaultView.frameElement;
// 		if (iframe && iframe.contentDocument) {
// 			const targetElement = iframe.contentDocument.getElementById(id);
// 			if (targetElement) {
// 				targetElement.scrollIntoView({ behavior: 'smooth' });
// 				return;
// 			} else {
// 				console.log(`Element with id ${id} not found`);
// 			}
// 		}
// 	}
// }

onMounted(() => {
	loadBook();
});

onBeforeUnmount(() => {
	if (book.value) {
		book.value.destroy();
	}
});
</script>

<style scoped>
#viewer {
	width: 100%;
	height: 100%;
}

.toc li {
	cursor: pointer;
	color: blue;
	text-decoration: underline;
	margin-bottom: 5px;
}

.toc li:hover {
	color: darkblue;
}
</style>
