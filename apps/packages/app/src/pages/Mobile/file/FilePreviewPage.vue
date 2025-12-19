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
				'preview-safe-area': menuStore.useSafeArea || $q.platform.is.android,
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
			<div class="content">
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
import { useDataStore } from '../../../stores/data';
import { onMounted, onUnmounted, ref, watch } from 'vue';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { useFilesStore, FilesIdType } from '../../../stores/files';
import { busOn } from '../../../utils/bus';
import { busOff } from '../../../utils/bus';
import { dataAPIs } from '../../../api';
import { useMenuStore } from '../../../stores/menu';

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

const cancel = () => {
	onDialogCancel();
};

const currentView = ref();

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
		const dataAPI = dataAPIs(newVal.driveType);
		if (
			newVal.type.toLowerCase() === 'text' ||
			newVal.type.toLowerCase() === 'txt' ||
			newVal.type.toLowerCase() === 'textImmutable'
		) {
			store.preview.isEditEnable =
				dataAPI.fileEditEnable && newVal.size <= dataAPI.fileEditLimitSize;
			return (currentView.value = FileEditor);
		} else {
			rightIcon.value = '';
			store.preview.isEditEnable = false;
			if (newVal.type == 'image' || newVal.type == 'audio') {
				isDark.value = true;
				if (!dataAPI.audioPlayEnable && newVal.type == 'audio') {
					return (currentView.value = FileUnavailable);
				}
				return (currentView.value = FileMediumPreview);
			} else if (newVal.type.toLowerCase() == 'pdf') {
				return (currentView.value = FilePDFPreview);
			}

			return (currentView.value = FileUnavailable);
		}
	},
	{
		deep: true
	}
);

watch(
	() => store.preview,
	() => {
		if (store.preview.isEditEnable) {
			if (store.preview.isEditing) {
				rightIcon.value = 'sym_r_save';
			} else {
				rightIcon.value = 'sym_r_edit_square';
			}
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
		onDialogCancel();
	});
});

onUnmounted(() => {
	store.preview.isShow = false;
	busOff('dialogDismiss');
});
</script>

<style scoped lang="scss">
#preview-previewer {
	position: relative;
	// background-color: rgba(0, 0, 0, 1);
	// background-color: $background-1;
	height: 100vh;
	max-height: 100vh !important;
}

#preview-previewer .preview {
	text-align: center;
	height: 100%;
}

#preview-previewer .content {
	width: 100%;
	height: calc(100% - 56px);
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
