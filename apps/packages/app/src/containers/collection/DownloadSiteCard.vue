<template>
	<BaseSiteCard :data="info">
		<template #action>
			<q-btn
				padding="6px"
				class="open-file-wrapper"
				v-if="data.is_exist"
				@click="openFile"
			>
				<q-icon
					name="sym_r_folder_open"
					:color="theme?.btnTextActiveColor"
					size="20px"
				/>
			</q-btn>
			<div
				v-else-if="DownloadStatusEnum.DOWNLOADING === data.download_status"
				class="row items-center justify-center"
			>
				<SpinnerLoading> </SpinnerLoading>
			</div>
			<q-btn
				:color="theme?.btnDefaultColor"
				padding="6px"
				v-else
				:loading="data.loading"
				@click="collectSiteStore.downloadFile(data)"
				:disable="data.disabled"
			>
				<q-icon
					name="sym_r_download"
					:color="theme?.btnTextDefaultColor"
					size="20px"
				/>
			</q-btn>

			<!-- <q-circular-progress
				v-show="DownloadStatusEnum.DOWNLOADING === data.download_status"
				show-value
				font-size="12px"
				:value="percent"
				size="32px"
				:thickness="0.22"
				color="teal"
				track-color="grey-3"
				class="q-ma-md"
			>
				{{ percent.toFixed(1) }}%
			</q-circular-progress> -->
		</template>
		<div
			class="text-overline text-ink-2 q-px-sm q-py-xs bg-background-hover subtitle-wrapper ellipsis"
		>
			<span class="capitalize-text">{{ data.file_type }}</span>
			<template v-if="data.resolution">
				<span>-</span>
				<span>{{ data.resolution }}</span>
			</template>
			<template v-if="data.filesize">
				<span>-</span>
				<span>{{ convertBytesString(data.filesize) }}</span>
			</template>
			<template v-if="data.ext">
				<span>-</span>
				<span class="uppercase-text">{{ data.ext }}</span>
			</template>
		</div>
	</BaseSiteCard>
</template>

<script setup lang="ts">
import BaseSiteCard from '../../components/collection/BaseSiteCard.vue';
import { computed, inject, ref, toRefs } from 'vue';
import { DownloadItem, DownloadStatusEnum } from 'src/types/commonApi';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { FilePath, useFilesStore } from 'src/stores/files';
import { DriveType } from 'src/utils/interface/files';
import { BaseSiteCardProps } from 'src/components/collection/collect';
import { convertBytesString } from 'src/utils/file';
import { getFileIcon } from '@bytetrade/core';
import SpinnerLoading from 'src/components/common/SpinnerLoading.vue';
import { COLLECT_THEME } from 'src/constant/provide';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
import { openUrl } from 'src/utils/bex/tabs';
import { useUserStore } from 'src/stores/user';
import { replaceOriginDomain } from 'src/utils/url2';
import { getApplication } from 'src/application/base';
interface Props {
	data: DownloadItem & BaseSiteCardProps['data'] & { disabled?: boolean };
}
const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);

const filesStore = useFilesStore();

const fileSavePathRef = ref<FilePath | undefined>(filesStore.currentPath[1]);

const collectSiteStore = useCollectSiteStore();
const props = withDefaults(defineProps<Props>(), {});
const setSelectPath = (fileSavePath: FilePath) => {
	fileSavePathRef.value = fileSavePath;
};
const info = computed(() => {
	return {
		id: props.data.id,
		title: props.data.file,
		url: props.data.url,
		icon: props.data.icon || fileIcon(props.data.ext)
	};
});

const openFile = () => {
	let url = '';
	const filesKey = 'files';

	if (getApplication().platform && getApplication().platform?.isClient) {
		const userStore = useUserStore();
		url = userStore.getModuleSever(
			filesKey,
			'https:',
			`/Files/Home/Downloads/${props.data.file}`
		);
	} else {
		const origin = replaceOriginDomain(location.origin, filesKey, true);
		url = `${origin}/Files/Home/Downloads/${props.data.file}`;
	}

	openUrl(url);
};

const fileIcon = (name: any) => {
	let src = '/img/file-';
	let folderSrc = '/img/file-other.svg';

	if (name?.split('.')?.length > 1) {
		src = src + getFileIcon(name) + '.svg';
	} else {
		src = folderSrc;
	}

	return src;
};

const percent = computed(() => {
	return props.data.percent || 0;
});
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
</style>
