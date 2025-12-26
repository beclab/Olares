<template>
	<q-dialog
		v-model="dialog"
		persistent
		:maximized="true"
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<q-card>
			<div id="previewer" class="bg-background-3">
				<header-bar>
					<div
						class="q-ml-md text-ink-1 text-subtitle3"
						:style="
							$q.platform.is.electron && $q.platform.is.mac
								? 'margin-left: 90px;'
								: ''
						"
					>
						{{ title }}
					</div>
					<template
						#info
						v-if="filesStore.previewItem[origin_id].type === 'image'"
					>
						<div
							class="q-px-md view-another-image text-caption text-ink-2 cursor-pointer q-mr-sm"
							@click="store.preview.fullSize = !store.preview.fullSize"
							style="-webkit-app-region: no-drag"
						>
							{{
								!store.preview.fullSize
									? $t('files.view_original_image') +
									  ' ' +
									  humanStorageSize(
											filesStore.previewItem[origin_id].size || 0
									  )
									: $t('files.view_preview_image')
							}}
						</div>
					</template>

					<template #actions>
						<action
							:disabled="store.loading"
							v-if="currentPermission === 'rw' && store.preview.isEditEnable"
							:icon="store.preview.isEditing ? 'save' : 'edit_square'"
							:label="$t('buttons.edit')"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="showOperation"
						/>

						<action
							:disabled="store.loading"
							v-if="currentPermission === 'rw' && store.user?.perm?.delete"
							icon="delete"
							:label="$t('buttons.delete')"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="deleteFile"
						/>

						<action
							:disabled="store.loading"
							v-if="store.user?.perm?.download"
							icon="browser_updated"
							:label="$t('buttons.download')"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="download"
						/>

						<action
							:disabled="store.loading"
							icon="close"
							:label="$t('buttons.close')"
							style="-webkit-app-region: no-drag"
							@action="close"
						/>
					</template>
				</header-bar>
				<div class="content bg-background-2">
					<BtLoading
						:show="true"
						v-if="store.loading"
						textColor="#4999ff"
						color="#4999ff"
						text=""
						backgroundColor="rgba(0, 0, 0, 0)"
					>
					</BtLoading>
					<template v-else>
						<component
							:is="currentView"
							:origin_id="origin_id"
							v-if="reloadFinished"
						></component>
					</template>

					<button
						@click="prev"
						:class="{ hidden: !hasPrevious }"
						:aria-label="$t('buttons.previous')"
						:title="$t('buttons.previous')"
						style="width: 32px; height: 32px"
					>
						<i class="material-icons">chevron_left</i>
					</button>
					<button
						@click="next"
						:class="{ hidden: !hasNext }"
						:aria-label="$t('buttons.next')"
						:title="$t('buttons.next')"
						style="width: 32px; height: 32px"
					>
						<i class="material-icons">chevron_right</i>
					</button>
				</div>
			</div>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import HeaderBar from '../../../components/files/header/HeaderBar.vue';
import Action from '../../../components/files/header/Action.vue';

import { useDataStore } from '../../../stores/data';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import FilePreview from './../common-files/FilePreview.vue';
import FileEditor from '../common-files/FileEditor.vue';
import FileEpubPreview from '../common-files/FileEpubPreview.vue';
import FileUnavailable from '../common-files/FileUnavailable.vue';
import FilePDFPreview from '../common-files/FilePDFPreview.vue';

import { onMounted, onUnmounted, ref, onBeforeMount } from 'vue';
import { useQuasar } from 'quasar';
import { computed } from 'vue';
import { watch } from 'vue';
import { nextTick } from 'process';
// import { dataAPIs } from '../../../api';
import { common } from './../../../api';
import { getAppPlatform } from '../../../application/platform';
import { useRoute } from 'vue-router';
import { useOperateinStore } from './../../../stores/operation';
import { OPERATE_ACTION } from '../../../utils/contact';
import { dataAPIs } from '../../../api';
import { format } from '../../../utils/format';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true,
		default: FilesIdType.PAGEID
	}
});

const dialog = ref(true);

const { humanStorageSize } = format;
const route = useRoute();

const store = useDataStore();
const filesStore = useFilesStore();
const operateinStore = useOperateinStore();

const title = ref(filesStore.previewItem[props.origin_id]?.name);

const currentView = ref();
const reloadFinished = ref(true);
const currentPermission = ref('rw');

const { t } = useI18n();

const $q = useQuasar();

const hasPrevious = computed(function () {
	if ((filesStore.currentFileList[props.origin_id]?.items?.length || 0) <= 1) {
		return false;
	}
	const items = filesStore.currentFileList[props.origin_id]?.items.filter(
		(e) => !e.isDir
	);
	const index = items?.findIndex(
		(e) => e.name == filesStore.previewItem[props.origin_id]?.name
	);

	return index && index > 0;
});

const hasNext = computed(function () {
	if ((filesStore.currentFileList[props.origin_id]?.items.length || 0) <= 1) {
		return false;
	}

	const items = filesStore.currentFileList[props.origin_id]?.items.filter(
		(e) => !e.isDir
	);

	if (!items) {
		return false;
	}

	const index = items.findIndex(
		(e) => e.name == filesStore.previewItem[props.origin_id]?.name
	);
	return index >= 0 && index < items.length - 1;
});

const deleteFile = () => {
	store.showHover({
		prompt: 'delete',
		confirm: () => {
			close();
		}
	});
};

const prev = async () => {
	const items = filesStore.currentFileList[props.origin_id]?.items.filter(
		(e) => !e.isDir
	);
	const index = items?.findIndex(
		(e) => e.name == filesStore.previewItem[props.origin_id]?.name
	);
	if (!index || !items || index < 1) {
		return;
	}

	let preItem = items[index - 1];
	filesStore.previewItem[props.origin_id] = preItem;
	await open(preItem);
};

const next = async () => {
	const items = filesStore.currentFileList[props.origin_id]?.items.filter(
		(e) => !e.isDir
	);
	if (!items) {
		return;
	}
	const index = items.findIndex(
		(e) => e.name == filesStore.previewItem[props.origin_id]?.name
	);
	if (index < 0 || index + 1 > items.length) {
		return;
	}
	let nextItem = items[index + 1];
	filesStore.previewItem[props.origin_id] = nextItem;
	await open(nextItem);
};

const open = async (item) => {
	await filesStore.openPreviewDialog(item);
};

watch(
	() => filesStore.previewItem[props.origin_id],
	async (newVal) => {
		if (newVal.type == undefined) {
			return null;
		}

		if (newVal.permission) {
			if (typeof newVal.permission == 'string') {
				currentPermission.value = newVal.permission;
			} else if (typeof newVal.permission == 'number') {
				currentPermission.value = newVal.permission >= 3 ? 'rw' : 'r';
			}
		}

		title.value = filesStore.previewItem[props.origin_id]?.name;
		currentView.value = undefined;
		reloadFinished.value = false;
		const dataAPI = dataAPIs(filesStore.previewItem[props.origin_id].driveType);
		if (
			newVal.type.toLowerCase() === 'text' ||
			newVal.type.toLowerCase() === 'txt' ||
			newVal.type.toLowerCase() === 'blob' ||
			newVal.type.toLowerCase() === 'textImmutable'
		) {
			store.preview.isEditEnable =
				dataAPI.fileEditEnable && newVal.size <= dataAPI.fileEditLimitSize;
			currentView.value = FileEditor;
		} else if (newVal.type.toLowerCase() === 'epub') {
			currentView.value = FileEpubPreview;
		} else {
			store.preview.isEditEnable = false;

			if (newVal.name) {
				title.value = newVal.name as string;
			}

			if (
				newVal.type == 'image' ||
				newVal.type == 'audio' ||
				newVal.type.toLowerCase() == 'pdf' ||
				(newVal.type.toLowerCase() == 'word' &&
					getSuffix(newVal.name) === 'docx') ||
				(newVal.type.toLowerCase() == 'excel' &&
					getSuffix(newVal.name) === 'xlsx')
			) {
				if (getAppPlatform().isPad && newVal.type.toLowerCase() == 'pdf') {
					currentView.value = FilePDFPreview;
				} else if (!dataAPI.audioPlayEnable && newVal.type == 'audio') {
					currentView.value = FileUnavailable;
				} else {
					currentView.value = FilePreview;
				}
			} else {
				currentView.value = FileUnavailable;
			}
		}
		nextTick(() => {
			reloadFinished.value = true;
		});
	},
	{
		deep: true
	}
);

const getSuffix = (fileName: string) => {
	let suffix = '';
	if (fileName) {
		const flieArr = fileName.split('.');
		suffix = flieArr[flieArr.length - 1];
	}

	return suffix;
};

const download = async (e: any) => {
	const driveType = common().formatUrltoDriveType(route.path);
	if (!driveType) {
		return;
	}
	operateinStore.handleFileOperate(
		props.origin_id,
		e,
		route,
		OPERATE_ACTION.DOWNLOAD,
		driveType,
		async (action: OPERATE_ACTION, data: any) => {
			console.log('handleFileOperate action', action);
			console.log('handleFileOperate data', data);
			notifySuccess(t('Download task added'));
		}
	);
};

const close = () => {
	dialog.value = false;
	filesStore.previewItem[props.origin_id] = {};
	setTimeout(() => {
		filesStore.isInPreview[props.origin_id] = '';
	}, 500);
};

const showOperation = () => {
	if (store.preview.isEditEnable) {
		if (!store.preview.isEditing) {
			store.preview.isEditing = true;
		} else {
			store.preview.isSaving = true;
		}
	}
};

const keydownEnter = (event: any) => {
	event.stopPropagation();
	switch (event.keyCode) {
		case 37:
			prev();
			break;

		case 39:
			next();
			break;

		default:
			break;
	}
};

onBeforeMount(() => {
	store.preview.isShow = true;
});

onMounted(async () => {
	window.addEventListener('keydown', keydownEnter);
});

onUnmounted(() => {
	window.removeEventListener('keydown', keydownEnter);
	store.preview.fullSize = false;
	store.preview.isShow = false;
});
</script>

<style scoped lang="scss">
.view-another-image {
	height: 24px;
	line-height: 24px;
	border-radius: 4px;
	text-align: center;
	background: $background-1;
}

#previewer .content > button {
	background-color: $dimmed-background;
}

.ipad-header {
	margin-top: env(safe-area-inset-top);
}

.android-pad-header {
	margin-top: 30px;
}

.android-pad-height {
	height: calc(100vh - 30px);
	margin-top: 30px;
}
</style>
