<template>
	<div class="files-list-root">
		<terminus-title-bar
			:title="title"
			:is-dark="isDark"
			:hook-back-action="true"
			@on-return-action="back"
		>
			<template v-slot:right>
				<div class="row items-center" v-if="origin === 'dialog'">
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_close"
						text-color="ink-2"
						@click="closeDialog"
					>
					</q-btn>
				</div>
			</template>
		</terminus-title-bar>
		<!-- :right-icon="rightIcon"
			@on-right-click="showOperation" -->
		<div class="content">
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
				from="sync"
				:lockTouch="true"
				@open-sync-page="openSyncPage"
				v-else
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useDataStore } from '../../../stores/data';

import { ref, onMounted, PropType } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import Errors from '../../Files/Errors.vue';
import ListingFiles from './ListingFiles.vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
// import DirOperationDialog from './DirOperationDialog.vue';
// import { useQuasar } from 'quasar';
import {
	useFilesStore,
	FilesIdType,
	PickType,
	FileItem
} from '../../../stores/files';
import { seahub } from './../../../api';
import { MenuItem } from '../../../utils/contact';
import { syncFilesFormat } from './../../../api';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: PickType.FOLDER
	}
});

const emits = defineEmits(['openSyncPage', 'close', 'back']);

const router = useRouter();
const store = useDataStore();
const filesStore = useFilesStore();
const error = ref<any>(null);
const route = useRoute();
// const $q = useQuasar();

const title = ref(route.query.name);

// const rightIcon = ref('sym_r_more_horiz');

const isDark = ref(false);
const origin_id = ref(props.origin_id || FilesIdType.PAGEID);

if (!props.origin_id) {
	filesStore.initIdState(origin_id.value);
}

// const showOperation = () => {
// $q.dialog({
// 	component: DirOperationDialog,
// 	componentProps: {}
// });
// };

const openSyncPage = () => {
	emits('openSyncPage');
};

const closeDialog = () => {
	emits('close');
};

const back = () => {
	if (props.origin_id) {
		emits('back', 'init');
	}
	router.back();
};

onMounted(async () => {
	filesStore.currentFileList[origin_id.value] = undefined;
	if (route.params.repo === MenuItem.MYLIBRARIES || origin_id.value) {
		const res = await seahub().fetchMineRepo();
		filesStore.currentFileList[origin_id.value] =
			await syncFilesFormat().formatSeahubRepos(route.params.repo, res);
	} else if (route.params.repo === MenuItem.SHAREDWITH) {
		const [sharedByMeRes, sharedToMeRes] = await Promise.all([
			seahub().fetchtosharedRepo(),
			seahub().fetchsharedRepo()
		]);

		const mineRes = await seahub().fetchMineRepo();
		const formatMineRes = await syncFilesFormat().formatSeahubRepos(
			route.params.repo,
			mineRes
		);

		const sharedByMeFormatted = sharedByMeRes
			.map((el) =>
				formatMineRes.items.find((item) => item.name === el.repo_name)
			)
			.filter((commonItem) => commonItem !== undefined) as FileItem[];

		const formatSharedToMeRes = await syncFilesFormat().formatSeahubRepos(
			route.params.repo,
			sharedToMeRes
		);

		filesStore.currentFileList[origin_id.value] = {
			...formatSharedToMeRes,
			items: [...formatMineRes.items, ...formatSharedToMeRes.items]
		};
	}
});
</script>

<style lang="scss" scoped>
.files-list-root {
	width: 100%;
	height: 100%;

	.content {
		width: 100%;
		height: calc(100% - 56px);
	}
}
</style>
