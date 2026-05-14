<template>
	<div class="column flex-gap-y-md">
		<BaseSiteCard
			v-for="(group, groupIndex) in groupedDataList"
			:key="groupIndex"
			:data="getInfo(group.selectedData)"
			class="full-width"
		>
			<template #status>
				<div
					class="absolute-full row justify-center items-center avatar-status img-mask"
					v-if="
						DownloadStatusEnum.DOWNLOADING ===
							group.selectedData?.download_status ||
						group.selectedData?.download_status === DownloadStatusEnum.PAUSED ||
						(group.selectedData?.download_status ===
							DownloadStatusEnum.WAITING &&
							!!group.selectedData?.percent)
					"
				>
					<q-circular-progress
						show-value
						font-size="8px"
						:value="group.selectedData?.percent || 0"
						size="30px"
						:thickness="0.26"
						color="ink-on-brand"
						track-color="background-alpha"
					>
						<img
							v-if="
								group.selectedData?.download_status ===
								DownloadStatusEnum.PAUSED
							"
							src="~/assets/plugin/paused.svg"
							alt="checked"
						/>
						<DotLoading
							v-else-if="
								group.selectedData?.download_status ===
									DownloadStatusEnum.WAITING && !!group.selectedData?.percent
							"
						></DotLoading>
						<span
							v-else
							class="text-overline text-white"
							style="font-size: 8px"
						>
							{{ round(group.selectedData?.percent || 0, 0) }}%
						</span>
					</q-circular-progress>
				</div>
				<div
					class="absolute-full row justify-center items-center avatar-status img-mask"
					v-else-if="
						group.selectedData?.download_status === DownloadStatusEnum.WAITING
					"
				>
					<SpinnerLoading type="overlay"> </SpinnerLoading>
				</div>
				<div
					class="absolute-full row justify-center items-center avatar-status img-mask"
					v-else-if="
						group.selectedData?.download_status ===
							DownloadStatusEnum.COMPLETE && !!group.selectedData?.is_exist
					"
				>
					<img src="~/assets/plugin/checked2.svg" alt="checked" />
				</div>

				<div
					class="absolute-full row justify-center items-center avatar-status img-mask"
					v-else-if="
						group.selectedData?.download_status === DownloadStatusEnum.ERROR
					"
				>
					<img src="~/assets/plugin/error.svg" alt="checked" />
				</div>

				<div
					class="absolute-full row justify-center items-center avatar-status img-mask"
					v-else-if="
						group.selectedData?.download_status === DownloadStatusEnum.CANCEL
					"
				>
					<img src="~/assets/plugin/cancel.svg" alt="checked" />
				</div>
			</template>
			<template #action>
				<DownloadAction
					v-if="
						group.selectedData?.task_id &&
						group.selectedData?.download_status !== DownloadStatusEnum.COMPLETE
					"
					:download_status="group.selectedData?.download_status"
					:task_id="group.selectedData?.task_id"
				></DownloadAction>
				<q-btn
					v-else-if="!group.selectedData?.is_exist"
					:color="theme?.btnDefaultColor"
					padding="6px"
					:loading="group.selectedData?.loading"
					@click="handleDownload(group.selectedData)"
					:disable="group.selectedData?.disabled"
				>
					<q-icon
						name="sym_r_download"
						:color="theme?.btnTextDefaultColor"
						size="20px"
					/>
				</q-btn>
				<q-btn
					v-else-if="
						group.selectedData?.download_status ===
							DownloadStatusEnum.COMPLETE &&
						isBoolean(group.selectedData?.is_exist) &&
						group.selectedData?.is_exist
					"
					padding="6px"
					class="open-file-wrapper"
					@click="openFile(group.selectedData.file)"
				>
					<q-icon
						name="sym_r_folder_open"
						:color="theme?.btnTextActiveColor"
						size="20px"
					/>
				</q-btn>
			</template>
			<QSelectStyleV4 class="q-mt-xs">
				<BtSelectV4
					borderless
					v-model="group.selectedData"
					:options="group.items"
					dense
					bg-color="background-hover"
					style="padding: 0"
					class="quality-select"
				>
					<template v-slot:selected>
						<div
							class="row items-center ellipsis text-overline text-ink-2 no-wrap"
						>
							<img
								style="height: 7px; margin-right: 4px"
								:src="fileIcon(group.selectedData)"
								alt=""
							/>
							<span class="ellipsis" style="flex: 1">{{
								getOptionLabel(group.selectedData)
							}}</span>
						</div>
					</template>
					<template v-slot:option="scope">
						<div
							class="row items-center justify-between q-pl-sm q-py-xs q-pr-xs q-mt-xs cursor-pointer select-option-item download-selected-hover"
							v-bind="scope.itemProps"
							:class="[
								group.selectedData?.id === scope.opt.id
									? 'download-selected-active text-orange-default'
									: 'text-ink-2'
							]"
						>
							<div class="row items-center">
								<img
									style="height: 9.6px; margin-right: 4px"
									:src="
										fileIcon(
											group.selectedData,
											group.selectedData?.id === scope.opt.id
										)
									"
									alt=""
								/>
								<span class="text-caption ellipsis">{{
									getOptionLabel(scope.opt)
								}}</span>
							</div>
							<q-icon name="check_circle" class="q-ml-sm" size="12px" />
						</div>
					</template>
				</BtSelectV4>
			</QSelectStyleV4>
		</BaseSiteCard>
	</div>
</template>

<script setup lang="ts">
import BaseSiteCard from '../../components/collection/BaseSiteCard.vue';
import { inject, ref, watch } from 'vue';
import { DownloadItem, DownloadStatusEnum } from 'src/types/commonApi';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { useFilesStore } from 'src/stores/files';
import { BaseSiteCardProps } from 'src/components/collection/collect';
import { getFileIcon } from '@bytetrade/core';
import SpinnerLoading from 'src/components/common/SpinnerLoading.vue';
import { COLLECT_THEME } from 'src/constant/provide';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
import { openUrl } from 'src/utils/bex/tabs';
import { useUserStore } from 'src/stores/user';
import { replaceOriginDomain } from 'src/utils/url2';
import { getApplication } from 'src/application/base';
import { round, isBoolean } from 'lodash';
import DownloadAction from './DownloadAction.vue';
import BtSelectV4 from 'src/components/settings/base/BtSelectV4.vue';
import QSelectStyleV4 from 'src/components/settings/base/QSelectStyleV4.vue';
import unknowIcon from 'src/assets/plugin/file-type-unknow.svg';
import videoIcon from 'src/assets/plugin/file-type-video.svg';
import audioIcon from 'src/assets/plugin/file-type-audio.svg';
import pdfIcon from 'src/assets/plugin/file-type-pdf.svg';
import ebookIcon from 'src/assets/plugin/file-type-ebook.svg';
import imageIcon from 'src/assets/plugin/file-type-image.svg';

import unknowIconHighlight from 'src/assets/plugin/file-type-unknow-highlight.svg';
import videoIconHighlight from 'src/assets/plugin/file-type-video-highlight.svg';
import audioIconHighlight from 'src/assets/plugin/file-type-audio-highlight.svg';
import pdfIconHighlight from 'src/assets/plugin/file-type-pdf-highlight.svg';
import ebookIconHighlight from 'src/assets/plugin/file-type-ebook-highlight.svg';
import imageIconHighlight from 'src/assets/plugin/file-type-image-highlight.svg';
import DotLoading from 'src/pages/Plugin/DotLoading.vue';

interface Props {
	dataList: (DownloadItem &
		BaseSiteCardProps['data'] & { disabled?: boolean })[];
}

interface GroupData {
	fileType: string;
	items: (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })[];
	selectedData:
		| (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })
		| undefined;
}

const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);
const filesStore = useFilesStore();
const collectSiteStore = useCollectSiteStore();
const props = withDefaults(defineProps<Props>(), {
	dataList: () => []
});

const parseResolution = (resolution: string): number => {
	if (!resolution) return 0;
	const match = resolution.match(/(\d+)/);
	return match ? parseInt(match[1]) : 0;
};

const selectBestQuality = (
	items: (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })[]
):
	| (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })
	| undefined => {
	if (!items || items.length === 0) {
		return undefined;
	}

	const sortedList = [...items];

	const hasCreatedTime = sortedList.some((item) => item.created_time);
	if (hasCreatedTime) {
		sortedList.sort((a, b) => {
			const timeA = a.created_time ? new Date(a.created_time).getTime() : 0;
			const timeB = b.created_time ? new Date(b.created_time).getTime() : 0;
			return timeB - timeA;
		});
	} else {
		const hasResolution = sortedList.some((item) => item.resolution);
		if (hasResolution) {
			sortedList.sort((a, b) => {
				const resA = parseResolution(a.resolution);
				const resB = parseResolution(b.resolution);
				return resB - resA;
			});
		} else {
			sortedList.sort((a, b) => {
				const sizeA = a.filesize || 0;
				const sizeB = b.filesize || 0;
				return sizeB - sizeA;
			});
		}
	}

	return sortedList[0];
};

const groupedDataList = ref<GroupData[]>([]);

const updateGroupedData = () => {
	const grouped = new Map<
		string,
		(DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })[]
	>();

	props.dataList.forEach((item) => {
		const fileType = item.file_type || 'other';
		if (!grouped.has(fileType)) {
			grouped.set(fileType, []);
		}
		grouped.get(fileType)!.push(item);
	});

	const newGroups = Array.from(grouped.entries()).map(([fileType, items]) => {
		const existingGroup = groupedDataList.value.find(
			(g) => g.fileType === fileType
		);

		let selectedData:
			| (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })
			| undefined;

		const userSelectedId = collectSiteStore.getUserSelection(fileType);
		if (userSelectedId) {
			selectedData = items.find((item) => item.id === userSelectedId);
		}

		if (!selectedData && existingGroup && existingGroup.selectedData) {
			const updatedItem = items.find(
				(item) => item.id === existingGroup.selectedData?.id
			);
			if (updatedItem) {
				selectedData = updatedItem;
			}
		}

		if (!selectedData) {
			selectedData = selectBestQuality(items);
		}

		return {
			fileType,
			items,
			selectedData
		};
	});

	groupedDataList.value = newGroups;
};

watch(
	() => props.dataList,
	() => {
		updateGroupedData();
	},
	{ immediate: true, deep: true }
);

watch(
	() =>
		groupedDataList.value.map((g) => ({
			fileType: g.fileType,
			selectedId: g.selectedData?.id
		})),
	(newSelections, oldSelections) => {
		if (!oldSelections || oldSelections.length === 0) {
			return;
		}

		newSelections.forEach((newSel, index) => {
			const oldSel = oldSelections[index];
			if (
				oldSel &&
				newSel.selectedId &&
				newSel.selectedId !== oldSel.selectedId
			) {
				collectSiteStore.saveUserSelection(newSel.fileType, newSel.selectedId);
			}
		});
	},
	{ deep: true }
);

const getOptionLabel = (item: any, fileType?: string): string => {
	if (!item) return '';

	const parts: string[] = [];

	if (item.resolution) {
		parts.push(item.resolution);
	} else if (item.file_type) {
		parts.push(item.file_type);
	}

	if (item.ext) {
		parts.push(item.ext);
	}

	return parts.length > 0 ? parts.join('-') : '';
};

const getInfo = (
	selectedData:
		| (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })
		| undefined
) => {
	return {
		...selectedData,
		id: selectedData?.id || '',
		title: selectedData?.file || selectedData?.title || '',
		url: selectedData?.url || '',
		icon: selectedData?.icon || fileAvatar(selectedData?.file)
	};
};

const handleDownload = (
	item:
		| (DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean })
		| undefined
) => {
	if (item) {
		collectSiteStore.downloadFile(item);
	}
};

const openFile = (file = '') => {
	let url = '';
	const filesKey = 'files';
	if (getApplication().platform && getApplication().platform?.isClient) {
		const userStore = useUserStore();
		url = userStore.getModuleSever(
			filesKey,
			'https:',
			`/Files/Home/Downloads/${file}`
		);
	} else {
		const origin = replaceOriginDomain(location.origin, filesKey, true);
		url = `${origin}/Files/Home/Downloads/${file}`;
	}

	openUrl(url);
};

const fileIcon = (item: any, highlight = false) => {
	const name = `.${item?.ext}`;
	let file_type = getFileIcon(name);
	if (item?.file_type && item.file_type !== file_type) {
		file_type = item.file_type;
	}
	switch (file_type) {
		case 'video':
			return highlight ? videoIconHighlight : videoIcon;
		case 'audio':
			return highlight ? audioIconHighlight : audioIcon;
		case 'pdf':
			return highlight ? pdfIconHighlight : pdfIcon;
		case 'ebook':
			return highlight ? ebookIconHighlight : ebookIcon;
		case 'image':
			return highlight ? imageIconHighlight : imageIcon;
		default:
			return highlight ? unknowIconHighlight : unknowIcon;
	}
};

const fileAvatar = (name: any) => {
	let src = '/img/file-';
	let folderSrc = '/img/file-other.svg';

	if (process.env.PLATFORM == 'BEX') {
		src = '/www/img/file-';
		folderSrc = '/www/img/file-other.svg';
	} else if (process.env.PLATFORM == 'DESKTOP') {
		src = './img/file-';
		folderSrc = './img/file-other.svg';
	}

	if (name?.split('.')?.length > 1) {
		src = src + getFileIcon(name) + '.svg';
	} else {
		src = folderSrc;
	}
	return src;
};
</script>

<style lang="scss" scoped>
.open-file-wrapper {
	border: 1px solid $btn-stroke;
}
.subtitle-wrapper {
	border-radius: 999px;
	overflow: hidden;
}
.uppercase-text {
	text-transform: uppercase;
}
.capitalize-text {
	text-transform: capitalize;
}
.quality-select {
	:deep(.q-field__control) {
		border-radius: 8px;
		min-height: 36px;
	}
	:deep(.q-field__native) {
		padding: 6px 12px;
	}
}
.file-type-prefix {
	font-weight: 600;
	letter-spacing: 0.5px;
	margin-right: 4px;
	color: $ink-1;
}
.img-mask {
	background: rgba(0, 0, 0, 0.6);
	backdrop-filter: blur(1px);
	pointer-events: none;
}
.download-selected-hover:hover {
	background-color: $btn-bg-hover;
	border-radius: 4px;
}
.download-selected-active {
	background-color: $btn-bg-hover;
	border-radius: 4px;
}
</style>
