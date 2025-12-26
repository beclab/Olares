<template>
	<div class="files-container row">
		<q-splitter
			v-model="splitterModel"
			unit="px"
			:limits="[240, 480]"
			style="height: 100%; width: 100%"
		>
			<template v-slot:before>
				<div class="files-left">
					<app-tree
						:app="app"
						:selected="selected"
						:chartNodes="chartNodes"
						@on-selected="onSelected"
						@update-path-node="updatePathNode"
					/>
				</div>
			</template>

			<template v-slot:after>
				<div class="files-right">
					<app-edit-file
						:fileInfo="fileStore.fileInfo"
						:isEditing="fileStore.isEditing"
						@on-save-file="onSaveFile"
						@change-code="changeCode"
					/>
				</div>
			</template>
		</q-splitter>
	</div>
</template>

<script lang="ts" setup>
import axios from 'axios';
import { useI18n } from 'vue-i18n';
import { useQuasar } from 'quasar';
import { useRouter, useRoute, onBeforeRouteLeave } from 'vue-router';
import { ref, watch, onMounted, onUnmounted, PropType } from 'vue';
import {
	ApplicationInfo,
	FilesSelectType,
	OPERATE_ACTION
} from '@apps/studio/src/types/core';

import { useDevelopingApps } from '../stores/app';
import { useFileStore } from './../stores/file';

import AppTree from './common/AppTree.vue';
import AppEditFile from './common/AppEditFile.vue';
import DialogConfirm from '../components/dialog/DialogConfirm.vue';
import { useBeforeUnloadSave } from 'src/composables/useBeforeUnloadSave';
import { resetStatus } from '../common/reset';

useBeforeUnloadSave(() => fileStore.isEditing);

const props = defineProps({
	app: {
		type: Object as PropType<ApplicationInfo>,
		required: true
	}
});

const store = useDevelopingApps();
const fileStore = useFileStore();

const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const route = useRoute();
const chartNodes = ref<any>([]);
const tempFile = ref('');
const selected = ref(props.app.chart);
const splitterModel = ref(240);

onBeforeRouteLeave((to, from, next) => {
	if (fileStore.isEditing) {
		$q.dialog({
			component: DialogConfirm,
			componentProps: {
				title: t('message.confirmation'),
				message: t('message.save_file')
			}
		})
			.onOk(async () => {
				await onSaveFile();
				resetStatus();
				next();
			})
			.onCancel(() => {
				resetStatus();
				next();
			});
	} else {
		next();
	}
});

watch(
	() => fileStore.fileInfo.code,
	(newVal) => {
		if (newVal !== tempFile.value) {
			fileStore.isEditing = true;
			fileStore.addEditStorage(route.path);
		} else {
			if (fileStore.isEditing) {
				fileStore.isEditing = false;
				fileStore.deleteEditStorage(route.path);
			}
		}
	}
);

watch(
	() => props.app.title,
	(newTitle) => {
		if (chartNodes.value && chartNodes.value.length > 0) {
			chartNodes.value[0].label = newTitle || props.app.appName;
		}
	}
);

const changeCode = (value) => {
	fileStore.fileInfo.code = value;
};

function onSaveFile() {
	if (!selected.value) return false;
	return fileStore.saveFile(selected.value);
}

const getChildren = (items: any) => {
	let children: FilesSelectType[] = [];

	for (let n in items) {
		const data = items[n];

		const selectData: FilesSelectType = {
			label: data.name,
			icon: data.isDir ? 'folder' : 'article',
			path: data.path,
			expandable: data.isDir,
			selectable: !data.isDir,
			children: data.isDir ? [{}] : null,
			isDir: data.isDir,
			lazy: data.isDir ? true : false
		};

		if (data.items) {
			selectData.children = getChildren(data.items);
		}

		children.push(selectData);
	}

	return children;
};

async function loadChart() {
	const res: any = await axios.get(store.url + '/api/files' + props.app.chart);

	const children = getChildren(res.items);

	chartNodes.value = [
		{
			label: props.app.title || props.app.appName,
			icon: 'folder',
			children: children,
			selectable: false,
			path: props.app.appName,
			isDir: true
		}
	];

	// loadFile();
}

const loadFile = () => {
	if (route.path.split('/')[3]) {
		return;
	}

	const defaultFile = chartNodes.value[0].children.find((item) => !item.isDir);

	updateRoute(defaultFile.path);
	fetchData(defaultFile.path);
};

const updatePathNode = async (item: {
	path: string;
	action: OPERATE_ACTION;
	name?: string;
}) => {
	const res: any = await axios.get(store.url + '/api/files/' + item.path);

	const children = getChildren(res.items);

	const replaceNodes = replaceObjectByPath(
		chartNodes.value,
		item.path,
		children
	);

	chartNodes.value = replaceNodes;

	if (
		item.action === OPERATE_ACTION.ADD_FILE ||
		item.action === OPERATE_ACTION.ADD_FOLDER
	) {
		onSelected(`${item.path}/${item.name}`);
	}
};

function replaceObjectByPath(array, targetPath, newObject) {
	return array.map((item) => {
		if (item.path === targetPath) {
			item.children = newObject;
			return item;
		}

		if (item.children && Array.isArray(item.children)) {
			return {
				...item,
				children: replaceObjectByPath(item.children, targetPath, newObject)
			};
		}

		return item;
	});
}

const updateRoute = (value) => {
	selected.value = value;
	router.push({
		path: `/app/${value}`
	});
};

const onSelected = async (value) => {
	if (fileStore.isEditing) {
		// checkFileSave(value);
		await onSaveFile();
		updateRoute(value);
		fetchData(value);
	} else {
		fetchData(value);
		updateRoute(value);
	}
};

const fetchData = async (value) => {
	await fileStore.getFile(value);

	fileStore.isEditing = false;
	tempFile.value = fileStore.currentFileData.content
		? fileStore.currentFileData.content
		: '';
	fileStore.fileInfo.code = fileStore.currentFileData.content
		? fileStore.currentFileData.content
		: '';
	fileStore.fileInfo.lang = fileStore.currentFileData.extension.startsWith('.')
		? fileStore.currentFileData.extension.slice(1)
		: fileStore.currentFileData.extension;
	fileStore.fileInfo.name = fileStore.currentFileData.name;

	console.log('fileInfo', fileStore.fileInfo);
};

const checkFileSave = (value) => {
	$q.dialog({
		component: DialogConfirm,
		componentProps: {
			title: t('message.confirmation'),
			message: t('message.save_file')
		}
	})
		.onOk(async () => {
			await onSaveFile();
			updateRoute(value);
			fetchData(value);
		})
		.onCancel(() => {
			updateRoute(value);
			fetchData(value);
		});
};

const handleKeyDown = async (event) => {
	if (
		(event.ctrlKey || event.metaKey) && // Ctrl (Windows/Linux) 或 Cmd (Mac)
		event.key === 's' &&
		!event.shiftKey &&
		!event.altKey
	) {
		event.preventDefault();
		await onSaveFile();
	}
};

onMounted(async () => {
	window.addEventListener('keydown', handleKeyDown);

	fileStore.filesEditStorage = [];
	fileStore.fileInfo = {
		code: '',
		name: '',
		lang: ''
	};

	const fileParamsPath = route.path.split('/').slice(2);
	if (fileParamsPath && fileParamsPath.length > 0) {
		updateRoute(fileParamsPath.join('/'));
	}

	await loadChart();
	if (selected.value && selected.value.split('/').length > 1) {
		await fetchData(selected.value);
	}
});

onUnmounted(() => {
	window.removeEventListener('keydown', handleKeyDown);
});
</script>

<style lang="scss" scoped>
.files-container {
	height: 100%;

	.files-left {
		width: 100%;
		background-color: $background-1;
		padding: 12px 4px;
	}
	.files-right {
		width: 100%;
		height: 100%;
		overflow: hidden;
		display: flex;
		align-items: flex-start;
		justify-content: center;
	}
}
</style>
