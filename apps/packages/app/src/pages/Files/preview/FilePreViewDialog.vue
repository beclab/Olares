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
						class="q-ml-md text-ink-1 text-subtitle3 title"
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
							:icon="
								store.preview.isEditing ? 'sym_r_save' : 'sym_r_edit_square'
							"
							:label="
								store.preview.isEditing
									? $t('buttons.save')
									: $t('buttons.edit')
							"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="showOperation"
						/>

						<action
							:disabled="store.loading"
							v-if="currentPermission === 'rw' && store.user?.perm?.delete"
							icon="sym_r_delete"
							:label="$t('buttons.delete')"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="deleteFile"
						/>

						<action
							:disabled="store.loading"
							v-if="store.user?.perm?.download"
							icon="sym_r_download"
							:label="$t('buttons.download')"
							style="-webkit-app-region: no-drag; margin-right: 4px"
							@action="download"
						/>

						<action
							:disabled="store.loading"
							icon="sym_r_close"
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
import UnsavedChangesDialog from '../../../components/files/prompts/UnsavedChangesDialog.vue';
import ArchivePreviewDialog from '../../../components/files/archive/ArchivePreviewDialog.vue';

import { onMounted, onUnmounted, ref, onBeforeMount } from 'vue';
import { useQuasar } from 'quasar';
import { computed } from 'vue';
import { watch } from 'vue';
import { nextTick } from 'vue';
// import { dataAPIs } from '../../../api';
import { common, isArchiveEnable } from './../../../api';
import { getAppPlatform } from '../../../application/platform';
import { useRoute } from 'vue-router';
import { useOperateinStore } from './../../../stores/operation';
import { OPERATE_ACTION } from '../../../utils/contact';
import { format } from '../../../utils/format';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import HotkeyManager from 'src/directives/hotkeyManager';
import { FILES_HOTKEY } from 'src/api/files/hotKeys';
import {
	isArchivePreviewSupported,
	isArchiveDriveTypeSupported
} from '../../../utils/interface/archive';
import {
	getPreviewPermission,
	isImagePreviewItem,
	resolveFilePreview
} from '../common-files/previewResolver';
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
const archivePreviewOpening = ref(false);

const { t } = useI18n();

const $q = useQuasar();

// Navigation is only available while previewing images, and it only
// switches between images (other file types are skipped).
const imageItems = computed(() =>
	(filesStore.currentFileList[props.origin_id]?.items || []).filter(
		isImagePreviewItem
	)
);

const currentImageIndex = computed(() =>
	imageItems.value.findIndex(
		(e) => e.name == filesStore.previewItem[props.origin_id]?.name
	)
);

const hasPrevious = computed(function () {
	if (store.preview.isEditing) {
		return false;
	}
	if (!isImagePreviewItem(filesStore.previewItem[props.origin_id])) {
		return false;
	}
	return currentImageIndex.value > 0;
});

const hasNext = computed(function () {
	if (store.preview.isEditing) {
		return false;
	}
	if (!isImagePreviewItem(filesStore.previewItem[props.origin_id])) {
		return false;
	}
	const index = currentImageIndex.value;
	return index >= 0 && index < imageItems.value.length - 1;
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
	if (!hasPrevious.value) {
		return;
	}
	const preItem = imageItems.value[currentImageIndex.value - 1];
	filesStore.previewItem[props.origin_id] = preItem;
	await open(preItem);
};

const next = async () => {
	if (!hasNext.value) {
		return;
	}
	const nextItem = imageItems.value[currentImageIndex.value + 1];
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
			const permission = getPreviewPermission(newVal.permission);
			if (permission) {
				currentPermission.value = permission;
			}
		}

		title.value = filesStore.previewItem[props.origin_id]?.name;
		if (newVal.name) {
			title.value = newVal.name as string;
		}
		if (
			isArchivePreviewSupported(newVal) &&
			isArchiveEnable() &&
			isArchiveDriveTypeSupported(newVal.driveType)
		) {
			openArchivePreview(newVal);
			return;
		}
		currentView.value = undefined;
		reloadFinished.value = false;
		const preview = resolveFilePreview(
			filesStore.previewItem[props.origin_id],
			{
				text: FileEditor,
				ebook: FileEpubPreview,
				pdf: FilePDFPreview,
				media: FilePreview,
				unavailable: FileUnavailable
			},
			{ usePdfComponent: getAppPlatform().isPad }
		);
		store.preview.isEditEnable = preview.isEditEnable;
		currentView.value = preview.component;
		nextTick(() => {
			reloadFinished.value = true;
		});
	},
	{
		deep: true
	}
);

const openArchivePreview = (item) => {
	if (archivePreviewOpening.value) {
		return;
	}
	archivePreviewOpening.value = true;
	dialog.value = false;
	$q.dialog({
		component: ArchivePreviewDialog,
		componentProps: {
			item,
			origin_id: props.origin_id,
			driveType: item.driveType
		}
	}).onDismiss(() => {
		filesStore.previewItem[props.origin_id] = {};
		filesStore.isInPreview[props.origin_id] = '';
		archivePreviewOpening.value = false;
	});
};

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
			notifySuccess(t('Download task added'));
		}
	);
};

const pendingClose = ref(false);

const doClose = () => {
	dialog.value = false;
	filesStore.previewItem[props.origin_id] = {};
	setTimeout(() => {
		filesStore.isInPreview[props.origin_id] = '';
	}, 500);
};

const close = () => {
	if (store.preview.isEditing && store.preview.isDirty) {
		$q.dialog({
			component: UnsavedChangesDialog
		}).onOk(({ action }: { action: 'save' | 'discard' }) => {
			if (action === 'save') {
				pendingClose.value = true;
				store.preview.isSaving = true;
			} else if (action === 'discard') {
				store.preview.isDirty = false;
				store.preview.isEditing = false;
				doClose();
			}
		});
		return;
	}
	doClose();
};

watch(
	() => store.preview.isSaving,
	(saving, prevSaving) => {
		if (!pendingClose.value) return;
		if (prevSaving && !saving) {
			if (!store.preview.isDirty) {
				doClose();
			}
			pendingClose.value = false;
		}
	}
);

const showOperation = () => {
	if (store.preview.isEditEnable) {
		if (!store.preview.isEditing) {
			store.preview.isEditing = true;
		} else {
			store.preview.isSaving = true;
		}
	}
};

onBeforeMount(() => {
	store.preview.isShow = true;
});

onMounted(async () => {
	HotkeyManager.registerHotkeys(
		{
			[FILES_HOTKEY.FILES.left]: () => {
				prev();
			},
			[FILES_HOTKEY.FILES.right]: () => {
				next();
			},
			[FILES_HOTKEY.FILES.down]: () => {
				close();
			}
		},
		['files']
	);
});

onUnmounted(() => {
	HotkeyManager.unregisterHotkeys(
		{
			[FILES_HOTKEY.FILES.left]: () => {
				prev();
			},
			[FILES_HOTKEY.FILES.right]: () => {
				next();
			},
			[FILES_HOTKEY.FILES.down]: () => {
				close();
			}
		},
		['files']
	);
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
	height: calc(100dvh - 30px);
	margin-top: 30px;
}

#previewer .title {
	max-width: calc(100% - 300px);
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}
</style>
