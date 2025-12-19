<template>
	<div class="files-root">
		<div class="content">
			<div class="share-info q-px-lg q-py-lg">
				<div class="row items-center">
					<q-avatar :size="`${20}px`">
						<TerminusAvatar :info="tokenStore.user" :size="20" />
					</q-avatar>
					<div class="text-body2 text-ink-2 q-ml-sm">
						{{ tokenStore.user?.olaresId?.split('@')[0] }}
					</div>
				</div>
				<template v-if="shareStore.share?.permission == SharePermission.Edit">
					<div class="row items-center q-mt-lg">
						<q-img src="/img/folder-default.svg" width="40px" />
						<div class="text-body2 text-ink-2">
							<div class="justify-between q-ml-md">
								<div class="text-subtitle1 text-ink-1">
									{{ shareStore.share?.name }}
								</div>
								<div class="text-body3 text-ink-2">
									{{
										t('files.Expiration date') +
										':' +
										formatFileModified(shareStore.share?.expire_time || '')
									}}
								</div>
							</div>
						</div>
					</div>
					<div
						class="text-body3 text-ink-3 q-mt-md"
						v-if="
							shareStore.share.upload_size_limit != undefined &&
							shareStore.share.upload_size_limit > 0
						"
					>
						{{
							t('share.The file size should be less than {size}', {
								size: getSuitableValue(
									`${shareStore.share.upload_size_limit}`,
									'disk'
								)
							})
						}}
					</div>
				</template>
				<template
					v-else-if="shareStore.share?.permission == SharePermission.UploadOnly"
				>
					<div class="row items-center q-mt-lg">
						<span class="text-ink1 text-h6 q-mr-xs">
							{{ t('Upload to') }}
						</span>
						<q-img src="/img/folder-default.svg" width="24px" />
						<span class="text-light-blue-default q-ml-xs text-h6">
							{{ shareStore.share?.name }}
						</span>
						<!-- <q-img src="/img/folder-default.svg" width="40px" />
						<div class="text-body2 text-ink-2">
							<div class="justify-between q-ml-md">
								<div class="text-subtitle1 text-ink-1">
									{{ shareStore.share?.name }}
								</div>
								<div class="text-body3 text-ink-2">
									{{
										t('files.Expiration date') +
										':' +
										formatFileModified(shareStore.share?.expire_time || '')
									}}
								</div>
							</div>
						</div> -->
					</div>
					<div
						class="text-body3 text-ink-3 q-mt-xs"
						v-if="
							shareStore.share.upload_size_limit != undefined &&
							shareStore.share.upload_size_limit > 0
						"
					>
						{{
							t('share.The file size should be less than {size}', {
								size: getSuitableValue(
									`${shareStore.share.upload_size_limit}`,
									'disk'
								)
							})
						}}
					</div>
				</template>
			</div>
			<template v-if="shareStore.share?.permission == SharePermission.Edit">
				<FilesPage :origin_id="origin_id" class="files-content" />
				<prompts-component :origin_id="origin_id" />
			</template>
			<template
				v-else-if="shareStore.share?.permission == SharePermission.UploadOnly"
			>
				<upload-only />
			</template>
		</div>
	</div>
</template>
<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import FilesPage from '../Files/FilesPage.vue';
import { DriveType } from 'src/utils/interface/files';
import { SharePermission } from 'src/utils/interface/share';
import { FilesIdType, useFilesStore } from 'src/stores/files';
import { useShareStore } from 'src/stores/share/share';
import PromptsComponent from 'src/components/files/prompts/PromptsComponent.vue';
import { useTokenStore } from 'src/stores/share/token';
import { formatFileModified } from '../../utils/file';
import UploadOnly from './UploadOnly.vue';
import { useI18n } from 'vue-i18n';

import { getSuitableValue } from 'src/utils/monitoring';

const id = FilesIdType.SHARE;
const filesStore = useFilesStore();
const shareStore = useShareStore();
const tokenStore = useTokenStore();
filesStore.initIdState(id);
const origin_id = ref(id);
const { t } = useI18n();

const initData = () => {
	let driveType = DriveType.PublicShare;

	const url = `/Share/${shareStore.path_id}/`;
	filesStore.setBrowserUrl(url, driveType, true, origin_id.value);
};

onMounted(() => {
	initData();
});
</script>

<style lang="scss" scoped>
.files-root {
	width: 100%;
	height: 100%;
	padding: 20px;

	.content {
		background-color: $background-1;
		width: 100%;
		height: 100%;
		border-radius: 12px;
		overflow: hidden;

		.share-info {
			width: 100%;
			height: 130px;
		}

		.files-content {
			height: calc(100% - 130px);
		}
	}
}
</style>
