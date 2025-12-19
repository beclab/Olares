<template>
	<div class="files-list-root">
		<terminus-title-bar
			v-if="selectIds === null"
			:title="filesStore.getCurrentRepo(route.path, origin_id)"
			:is-dark="isDark"
			:hook-back-action="true"
			@on-return-action="onReturnAction"
		>
			<template v-slot:right>
				<div class="row items-center" v-if="origin_id">
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_close"
						text-color="ink-2"
						@click="closeDialog"
					>
					</q-btn>
				</div>
				<div class="row items-center" v-else>
					<q-btn
						v-if="showFloatingButton"
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_checklist"
						text-color="ink-2"
						@click="intoCheckedMode"
					>
					</q-btn>
					<q-btn
						v-if="isShareRoot"
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border q-mr-xs"
						icon="sym_r_tune"
						text-color="ink-2"
						@click="showFilter"
					/>
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_more_horiz"
						text-color="ink-2"
					>
						<PopupMenu :origin_id="origin_id" @handleEvent="handleEvent" />
					</q-btn>
				</div>
			</template>
		</terminus-title-bar>
		<terminus-select-header
			v-else
			:selectIds="selectIds"
			@handle-close="handleClose"
			@handle-select-all="handleSelectAll"
			@handle-remove="removeFiles"
		/>

		<div class="content" :class="filesStore.isShard ? 'shard' : ''">
			<div
				v-if="store.loading"
				style="width: 100%; height: 100%"
				class="row items-center justify-center"
			>
				<q-spinner-dots color="primary" size="3em" />
			</div>
			<errors v-else-if="error" :errorCode="error.status" />
			<listing-files
				:origin_id="origin_id"
				:selectType="selectType"
				:lock-touch="isShareRoot"
				ref="FilesListRef"
				@show-select-mode="showSelectMode"
				v-else
			/>
		</div>

		<listing-files-footer
			v-if="showSharedButtons"
			:origin_id="origin_id"
			:selectType="selectType"
			@onSubmit="onSubmit"
		/>
		<add-files v-if="showFloatingButton" @addFile="addFile" />
	</div>
</template>

<script lang="ts" setup>
import { ref, onMounted, watch, computed, PropType } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import Errors from '../../Files/Errors.vue';
import ListingFiles from './ListingFiles.vue';
import TerminusSelectHeader from '../../../components/common/TerminusSelectHeader.vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import DirOperationDialog from './DirOperationDialog.vue';
import ListingFilesFooter from './ListingFilesFooter.vue';
import { useQuasar } from 'quasar';
import { useFilesStore, FilesIdType, PickType } from '../../../stores/files';
import { useDataStore } from '../../../stores/data';
import { commonV2 } from '../../../api';
import { FilesSortType, OPERATE_ACTION } from '../../../utils/contact';
import { hideHeaderOpt } from '../../../utils/file';
import { useOperateinStore } from '../../../stores/operation';
import PopupMenu from './PopupMenu.vue';
import AddFiles from './AddFiles.vue';
import { useI18n } from 'vue-i18n';
import FilterMobileView from '../../../components/files/share/FilterMobileView.vue';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: ''
	}
});

const emits = defineEmits(['back', 'close', 'onSubmit']);

const route = useRoute();
const router = useRouter();
const store = useDataStore();
const filesStore = useFilesStore();
const operateinStore = useOperateinStore();
const error = ref<any>(null);
const origin_id = ref(props.origin_id || FilesIdType.PAGEID);

const $q = useQuasar();
const isDark = ref(false);

const selectIds = ref(null);
const FilesListRef = ref();
const { t } = useI18n();

const closeDialog = () => {
	emits('close');
};
const onSubmit = (value) => {
	emits('onSubmit', value);
};

const showSharedButtons = computed(() => {
	if (!hideHeaderOpt(route.path)) {
		return false;
	}
	return filesStore.isShard || props.selectType.length > 0;
});

const showFloatingButton = computed(() => {
	if (filesStore.isShard) {
		return false;
	}

	if (selectIds.value) {
		return false;
	}

	if (props.selectType) {
		return false;
	}

	if (isShareRoot.value) {
		return false;
	}

	if (!hideHeaderOpt(route.path)) {
		return false;
	}

	return true;
});

const isShareRoot = computed(() => {
	return commonV2.isShareRootPage(route.path);
});

const showSelectMode = (value) => {
	selectIds.value = value;
};

const handleClose = () => {
	if (FilesListRef.value) {
		FilesListRef.value.handleClose();
	}
};

const handleSelectAll = () => {
	if (FilesListRef.value) {
		FilesListRef.value.handleSelectAll();
	}
};

const intoCheckedMode = () => {
	if (filesStore.isShard) {
		return false;
	}
	if (FilesListRef.value) {
		FilesListRef.value.intoCheckedMode();
	}
};

const showFilter = () => {
	$q.dialog({
		component: FilterMobileView,
		componentProps: {}
	}).onOk((value) => {});
};

const onReturnAction = () => {
	console.log('onReturnAction', filesStore.backStack);
	if (props.origin_id) {
		if (filesStore.backStack[props.origin_id].length === 1) {
			emits('back');
		}
	} else {
		router.go(-1);
	}
	filesStore.back(props.origin_id);
};

const addFile = () => {
	$q.dialog({
		component: DirOperationDialog,
		componentProps: {}
	}).onOk((value) => {
		console.log('addFileaddFile', value);
		if (value === 'uploadFiles') {
			// uploadFiles();
			filesStore.selectUploadFiles();
		} else if (value === 'createFolder') {
			createFolder();
		} else if (value === 'uploadImgOrVideo') {
			console.log(value);
			filesStore.selectUploadFiles(true);
		}
	});
};

const createFolder = () => {
	operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.CREATE_FOLDER,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

const removeFiles = async () => {
	const ids = selectIds.value.map((id) => Number(id.slice(0, id.indexOf('_'))));
	console.log('ids', ids);
	filesStore.selected[origin_id.value] = ids;
	await operateinStore.handleFileOperate(
		origin_id.value,
		null,
		route,
		OPERATE_ACTION.DELETE,
		filesStore.activeMenu(origin_id.value).driveType,
		async () => {}
	);
};

// const enterHistory = () => {
// 	router.push('/transfer/history');
// };

const handleEvent = (value: string) => {
	switch (value) {
		case 'mosaic':
			switchView(value);
			break;

		case 'list':
			switchView(value);
			break;

		case 'name':
			fileSort(FilesSortType.NAME);
			break;

		case 'type':
			fileSort(FilesSortType.TYPE);
			break;

		case 'file_modified':
			fileSort(FilesSortType.Modified);
			break;

		case 'size':
			fileSort(FilesSortType.SIZE);
			break;

		default:
			break;
	}
};

const switchView = async (type: string) => {
	if (type === store.user.viewMode) {
		return false;
	}
	const data = {
		id: store.user.id,
		viewMode: type
	};
	// users.update(data, ['viewMode']).catch();
	store.updateUser(data);
};

const fileSort = (sort: FilesSortType) => {
	if (filesStore.activeSort[origin_id.value].by == sort) {
		filesStore.updateActiveSort(
			sort,
			!filesStore.activeSort[origin_id.value].asc
		);
	} else {
		filesStore.updateActiveSort(sort, true);
	}
};

watch(
	() => operateinStore.executing,
	(newVal, oldVal) => {
		if (newVal === oldVal) {
			return false;
		}

		if (!newVal) {
			if (FilesListRef.value && selectIds.value) {
				FilesListRef.value.handleRemove();
			}
		}
	}
);

onMounted(() => {
	filesStore.selected[origin_id.value] = [];
	filesStore.refreshPathItems(origin_id.value);
});
</script>

<style lang="scss" scoped>
.files-list-root {
	width: 100vw;
	height: 100%;

	.content {
		width: 100%;
		height: calc(100% - 56px);
		&.shard {
			padding-bottom: 64px;
		}
	}
}
</style>
