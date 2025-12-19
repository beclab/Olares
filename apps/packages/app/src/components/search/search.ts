import { onMounted, onUnmounted, ref, watch } from 'vue';
import { debounce } from 'quasar';
import { TextSearchItem, ServiceType } from 'src/utils/interface/search';
import { useSearchStore } from 'src/stores/search';
import { useI18n } from 'vue-i18n';

import { searchCancel } from 'src/api/common/search';

export function useSearch(
	serviceType: ServiceType,
	open: (file: TextSearchItem) => void,
	emits: any,
	handSearchFiles: string | undefined
) {
	const searchFiles = ref('');

	const fileData = ref<TextSearchItem[] | undefined>();

	const pattern = new RegExp('[\u4E00-\u9FA5]+');

	const searchStore = useSearchStore();

	const hasMore = ref(false);

	const activeItem = ref<number | string | undefined>();

	const scrollArea = ref();

	const scrollAreaBG = ref();

	const { t } = useI18n();

	const filesRef = ref();

	watch(
		() => searchFiles.value,
		debounce(async (newVal: string | undefined) => {
			if (!newVal) {
				return (fileData.value = []);
			}
			if (!pattern.test(newVal) && newVal.length <= 2) {
				return false;
			}

			fileData.value = await getContent(newVal);
			hasMore.value = fileData.value.length >= searchStore.limit;
		}, 600)
	);

	const keydownEnter = (event: any) => {
		if (event.keyCode === 8 && !searchFiles.value) {
			return goBack();
		}

		if (!fileData.value) return false;
		const index = fileData.value.findIndex(
			(item: { id: number | string }) => item.id === activeItem.value
		);
		const selectItem = fileData.value[index];
		const upIndex = index - 1;
		const downIndex = index + 1;

		switch (event.keyCode) {
			case 38:
				if (upIndex >= 0) {
					activeItem.value = fileData.value[upIndex].id;
					refreshScroll(upIndex);
					chooseFile(upIndex);
				}
				break;

			case 40:
				if (downIndex <= fileData.value.length - 1) {
					activeItem.value = fileData.value[downIndex].id;
					refreshScroll(downIndex);
					chooseFile(downIndex);
				}
				break;

			case 13:
				open(selectItem);
				break;
		}
	};

	const refreshScroll = (index: number) => {
		const duration = 0;

		const scrollAreaHeight = scrollAreaBG.value.offsetHeight - 20;

		const displayNumbers = Math.floor((scrollAreaHeight || 0) / 60);

		if (index < displayNumbers) {
			scrollArea.value &&
				scrollArea.value?.$refs?.scrollRef?.setScrollPosition(
					'vertical',
					0,
					duration
				);
		} else if (index + displayNumbers > fileData.value!.length) {
			const scrollTarget =
				scrollArea.value?.$refs?.scrollRef?.getScrollTarget();
			scrollArea.value &&
				scrollArea.value?.$refs?.scrollRef?.setScrollPosition(
					'vertical',
					scrollTarget.scrollHeight,
					duration
				);
		} else {
			scrollArea.value &&
				scrollArea.value?.$refs?.scrollRef?.setScrollPosition(
					'vertical',
					60 * (index - displayNumbers + 2),
					duration
				);
		}
	};

	const chooseFile = (index: number) => {
		if (!fileData.value) {
			return;
		}
		activeItem.value = fileData.value[index].id;
	};

	const goBack = () => {
		searchStore.cancelSearch();
		emits('goBack');
	};

	const getContent = (query?: string) => {
		return searchStore.getContent(query, serviceType);
	};

	const handleClear = () => {
		searchFiles.value = '';
	};

	const getMore = async (index: number, done: any) => {
		const results = await searchStore.searchMore(fileData.value?.length || 0);
		hasMore.value = results.length >= searchStore.limit;

		if (fileData.value) {
			fileData.value = fileData.value?.concat(results);
		} else {
			fileData.value = results;
		}

		if (done) {
			done(!hasMore.value);
		}
	};

	onMounted(async () => {
		filesRef.value && filesRef.value.focus();
		if (handSearchFiles) searchFiles.value = handSearchFiles;
		window.addEventListener('keydown', keydownEnter);
	});

	onUnmounted(() => {
		window.removeEventListener('keydown', keydownEnter);
	});

	return {
		keydownEnter,
		searchFiles,
		filesRef,
		goBack,
		handleClear,
		getMore,
		fileData,
		t,
		activeItem,
		searchStore,
		pattern,
		scrollArea,
		scrollAreaBG,
		hasMore,
		chooseFile,
		refreshScroll
	};
}
