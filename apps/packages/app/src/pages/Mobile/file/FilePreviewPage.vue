<template>
	<q-dialog
		ref="dialogRef"
		class="preview-dialog"
		persistent
		:maximized="maximizedToggle"
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<div
			id="preview-previewer"
			:class="{
				'preview-safe-area': menuStore.useSafeArea && $q.platform.is.ios,
				'bg-dark-page': isDark,
				'bg-white-page': !isDark
			}"
		>
			<terminus-title-bar
				:title="title"
				:is-dark="isDark"
				:right-icon="rightIcon"
				:hook-back-action="true"
				@on-right-click="showOperation"
				@on-return-action="cancel"
			/>
			<div class="content" v-touch-swipe.mouse.horizontal="onSwipe">
				<div
					v-if="store.loading"
					style="width: 100%; height: 100%"
					class="row items-center justify-center"
				>
					<q-spinner-dots color="primary" size="3em" />
				</div>
				<component
					v-else-if="currentView"
					:is="currentView"
					:origin_id="origin_id"
				/>

				<button
					class="preview-nav preview-nav--prev"
					:class="{ hidden: !hasPrevious }"
					:aria-label="$t('buttons.previous')"
					@click="prev"
				>
					<i class="material-icons">chevron_left</i>
				</button>
				<button
					class="preview-nav preview-nav--next"
					:class="{ hidden: !hasNext }"
					:aria-label="$t('buttons.next')"
					@click="next"
				>
					<i class="material-icons">chevron_right</i>
				</button>
			</div>
		</div>
	</q-dialog>
</template>

<script setup lang="ts">
import FileEditor from '../../Files/common-files/FileEditor.vue';
import FileUnavailable from '../../Files/common-files/FileUnavailable.vue';
import FileMediumPreview from './FileMediumPreview.vue';
import FilePDFPreview from './FilePDFPreview.vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import UnsavedChangesDialog from '../../../components/files/prompts/UnsavedChangesDialog.vue';
import ArchivePreviewDialog from '../../../components/files/archive/ArchivePreviewDialog.vue';
import { useDataStore } from '../../../stores/data';
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { busOn } from '../../../utils/bus';
import { busOff } from '../../../utils/bus';
import { useMenuStore } from '../../../stores/menu';
import FileEpubPreview from '../../Files/common-files/FileEpubPreview.vue';
import { PreviewType } from '@bytetrade/core';
import {
	isImagePreviewItem,
	resolveFilePreview
} from '../../Files/common-files/previewResolver';
import {
	isArchivePreviewSupported,
	isArchiveDriveTypeSupported
} from '../../../utils/interface/archive';
import { computed } from 'vue';
import { isArchiveEnable } from '../../../api';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const title = ref('');

const isDark = ref(false);

const rightIcon = ref('sym_r_more_horiz');

const store = useDataStore();
const filesStore = useFilesStore();
const menuStore = useMenuStore();
const $q = useQuasar();

const maximizedToggle = ref(true);

const { dialogRef, onDialogCancel } = useDialogPluginComponent();

const pendingClose = ref(false);

const cancel = () => {
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
				onDialogCancel();
			}
		});
		return;
	}
	onDialogCancel();
};

watch(
	() => store.preview.isSaving,
	(saving, prevSaving) => {
		if (!pendingClose.value) return;
		if (prevSaving && !saving) {
			if (!store.preview.isDirty) {
				onDialogCancel();
			}
			pendingClose.value = false;
		}
	}
);

const currentView = ref();
const archivePreviewOpening = ref(false);

watch(
	() => filesStore.previewItem[props.origin_id],
	(newVal) => {
		if (newVal.type == undefined) {
			return null;
		}

		currentView.value = undefined;

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
		rightIcon.value = '';
		const preview = resolveFilePreview(
			filesStore.previewItem[props.origin_id],
			{
				text: FileEditor,
				ebook: FileEpubPreview,
				pdf: FilePDFPreview,
				media: FileMediumPreview,
				unavailable: FileUnavailable
			},
			{
				usePdfComponent: true,
				darkTypes: [PreviewType.IMAGE, PreviewType.AUDIO]
			}
		);
		store.preview.isEditEnable = preview.isEditEnable;
		if (store.preview.isEditEnable) {
			rightIcon.value = 'sym_r_edit_square';
		}
		isDark.value = preview.isDark;
		currentView.value = preview.component;
	},
	{
		deep: true
	}
);

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

const hasPrevious = computed(() => {
	if (store.preview.isEditing) {
		return false;
	}
	if (!isImagePreviewItem(filesStore.previewItem[props.origin_id])) {
		return false;
	}
	return currentImageIndex.value > 0;
});

const hasNext = computed(() => {
	if (store.preview.isEditing) {
		return false;
	}
	if (!isImagePreviewItem(filesStore.previewItem[props.origin_id])) {
		return false;
	}
	const index = currentImageIndex.value;
	return index >= 0 && index < imageItems.value.length - 1;
});

const onSwipe = ({ direction }: { direction: string }) => {
	if (!isImagePreviewItem(filesStore.previewItem[props.origin_id])) {
		return;
	}
	if (direction === 'left') {
		next();
	} else if (direction === 'right') {
		prev();
	}
};

const prev = async () => {
	if (!hasPrevious.value) {
		return;
	}
	const preItem = imageItems.value[currentImageIndex.value - 1];
	filesStore.previewItem[props.origin_id] = preItem;
	await filesStore.openPreviewDialog(preItem);
};

const next = async () => {
	if (!hasNext.value) {
		return;
	}
	const nextItem = imageItems.value[currentImageIndex.value + 1];
	filesStore.previewItem[props.origin_id] = nextItem;
	await filesStore.openPreviewDialog(nextItem);
};

const openArchivePreview = (item) => {
	if (archivePreviewOpening.value) {
		return;
	}
	archivePreviewOpening.value = true;
	onDialogCancel();
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

watch(
	() => store.preview,
	() => {
		if (store.preview.isEditEnable) {
			if (store.preview.isEditing) {
				rightIcon.value = 'sym_r_save';
			} else {
				rightIcon.value = 'sym_r_edit_square';
			}
		} else {
			rightIcon.value = '';
		}
	},
	{ deep: true }
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

onMounted(() => {
	isDark.value = false;
	store.preview.isShow = true;
	busOn('dialogDismiss', () => {
		cancel();
	});
});

onUnmounted(() => {
	store.preview.isShow = false;
	store.preview.isSaving = false;
	store.preview.isEditing = false;
	store.preview.isDirty = false;
	busOff('dialogDismiss');
});
</script>

<style scoped lang="scss">
#preview-previewer {
	position: relative;
	// background-color: rgba(0, 0, 0, 1);
	// background-color: $background-1;
	height: 100dvh;
	max-height: 100dvh !important;
}

#preview-previewer .preview {
	text-align: center;
	height: 100%;
}

#preview-previewer .content {
	width: 100%;
	height: calc(100% - 56px);
	position: relative;
}

.preview-nav {
	position: absolute;
	top: 50%;
	transform: translateY(-50%);
	z-index: 10;
	width: 36px;
	height: 36px;
	display: flex;
	align-items: center;
	justify-content: center;
	border: none;
	border-radius: 50%;
	color: #fff;
	background-color: rgba(0, 0, 0, 0.4);
	cursor: pointer;

	&--prev {
		left: 12px;
	}

	&--next {
		right: 12px;
	}

	&.hidden {
		display: none;
	}

	i {
		font-size: 24px;
	}
}
.preview-list-root {
	width: 100%;
	// height: 100%;
	position: relative;
	height: 100vh;
	max-height: 100vh !important;
	background-color: $background-1;

	.content {
		width: 100%;
		height: calc(100% - 56px);
	}

	.content-full {
		width: 100%;
		height: 100%;
	}
}

.preview-safe-area {
	padding-top: env(safe-area-inset-top);
	padding-bottom: env(safe-area-inset-bottom);
}

.bg-dark-page {
	background-color: $shadow-color;
}

.bg-white-page {
	background-color: $background-1;
}

.preview-dialog {
	::v-deep(.q-dialog__inner) {
		padding-top: 0px !important;
		padding-bottom: 0px !important;
	}
}
</style>
