<template>
	<bt-custom-dialog
		ref="customRef"
		:title="
			selectType === PickType.FOLDER
				? t('files.select_directory')
				: t('files.select_file')
		"
		:skip="t('files.add_directory')"
		:okLoading="loading ? t('loading') : false"
		:cancel="t('cancel')"
		:ok="t('confirm')"
		size="large"
		@onSubmit="submit"
		@onSkip="createDir"
	>
		<div class="pick-folder row items-center justify-between">
			<div class="pick-menu">
				<bt-menu
					:items="filesStore.menu[origin_id]"
					:modelValue="filesStore.activeMenu(origin_id).id"
					:sameActiveable="false"
					@select="selectHandler"
					style="width: 180px"
					active-class="text-subtitle2 bg-yellow-soft text-ink-1"
					size="sm"
				>
				</bt-menu>
			</div>
			<div class="pick-content">
				<dialog-header :origin_id="origin_id" />
				<dialog-listing
					:origin_id="origin_id"
					:selectType="selectType"
					@selectedFile="getSelectFiles"
				/>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, PropType } from 'vue';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

import { DriveType } from '../../utils/interface/files';
import { useFilesStore, PickType } from './../../stores/files';

import DialogHeader from './DialogHeader.vue';
import DialogListing from './DialogListing.vue';
// import TerminusDialogFooter from './../../components/common/TerminusDialogFooter.vue';
import NewDir from './../../components/files/prompts/NewDir.vue';
import { formatFilePath } from '../../constant';
import { filesIsV2 } from 'src/api';

// const { dialogRef, onDialogCancel, onDialogOK } = useDialogPluginComponent();

const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: PickType.FOLDER
	},

	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			if (filesIsV2()) {
				return [
					DriveType.Drive,
					// 0730 hide sync
					// DriveType.Sync,
					DriveType.External,
					DriveType.Cache,
					DriveType.Data
				];
			}
			return [
				DriveType.Drive,
				DriveType.External,
				DriveType.Sync,
				DriveType.Cache,
				DriveType.Data
			];
		}
	}
});

const emits = defineEmits(['onSubmit']);

const customRef = ref();
const { t } = useI18n();
const $q = useQuasar();
// const show = ref(true);
const loading = ref(false);
const filesStore = useFilesStore();

const initPath = ref('/Files/Home/');

const selectHandler = async (value) => {
	const path = await filesStore.formatRepotoPath(value.item, props.origin_id);

	const splitUrl = path.split('?');
	let param = '';
	if (splitUrl.length > 1) {
		param = '?' + splitUrl[1];
	}

	filesStore.setFilePath(
		{
			path: splitUrl[0],
			isDir: true,
			driveType: value.item.driveType,
			param
		},
		false,
		true,
		props.origin_id
	);
};

const getSelectFiles = () => {
	const selectedFile: any[] = [];
	for (const item of filesStore.selected[props.origin_id]) {
		const selected = filesStore.getTargetFileItem(item, props.origin_id);
		selectedFile.push(selected);
	}

	return selectedFile;
};

const submit = async () => {
	loading.value = true;

	if (props.selectType === PickType.FILE) {
		if (
			filesStore.selected[props.origin_id] &&
			filesStore.selected[props.origin_id].length > 0
		) {
			emits('onSubmit', getSelectFiles());
			// onDialogOK(getSelectFiles());
			customRef.value.onDialogOK(getSelectFiles());
		} else {
			BtNotify.show({
				type: NotifyDefinedType.WARNING,
				message: 'Please select a file'
			});
			loading.value = false;
			return false;
		}
	} else {
		const formatData = formatFilePath(filesStore.currentPath[props.origin_id]);
		emits('onSubmit', formatData);
		customRef.value.onDialogOK(formatData);
		// onDialogOK(filesStore.currentPath[props.origin_id]);
	}

	loading.value = false;
};

const createDir = () => {
	$q.dialog({
		component: NewDir,
		componentProps: {
			origin_id: props.origin_id
		}
	});
};

// const hide = () => {
// 	console.log('hide');
// 	filesStore.removeIdState(props.origin_id);
// };

onMounted(async () => {
	filesStore.setFilePath(
		{
			path: initPath.value,
			isDir: true,
			driveType: DriveType.Drive,
			param: ''
		},
		false,
		true,
		props.origin_id
	);
	await filesStore.getMenu(props.origins, props.origin_id);
});
</script>

<style lang="scss" scoped>
.pick-folder {
	width: 100%;
	height: 376px;
	max-width: 80vw;
	border-radius: 8px;
	padding: 0;
	overflow: hidden;
	border: 1px solid $separator;

	.pick-menu {
		width: 180px;
		height: 100%;
		border-right: 1px solid $separator;
		overflow-y: auto;
		overflow-x: hidden;
		&::-webkit-scrollbar {
			width: 0px;
		}
	}
	.pick-content {
		height: 100%;
		flex: 1;
	}
	.pick-content {
		width: 100%;
	}

	.but-cancel {
		border-radius: 8px;
		border: 1px solid $btn-stroke;
	}
}
</style>
