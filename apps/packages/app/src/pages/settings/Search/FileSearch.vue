<template>
	<page-title-component :show-back="true" :title="t('File Search')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature
			image="settings/imgs/root/search.svg"
			:label="t('Search Index')"
			:description="t('Rebuilds search index using the latest settings.')"
		>
			<template v-slot:status>
				<div
					v-if="taskStatus === 'running'"
					class="status-button q-ml-sm text-caption text-info bg-blue-alpha"
				>
					{{ t('In Progress') }}
				</div>
				<div
					v-else
					class="status-button q-ml-sm text-caption text-green bg-green-alpha"
				>
					{{ t('Completed') }}
				</div>
			</template>
			<template v-slot:button>
				<q-btn
					dense
					class="rebuild-button q-px-md q-py-sm text-body3 text-ink-2"
					:label="t('Rebuild')"
					no-caps
					@click="rebuildTask"
				/>
			</template>
		</app-menu-feature>

		<module-title class="q-mt-xl">
			<div class="row justify-start items-center">
				<div>{{ t('Excluded Files') }}</div>
				<q-icon size="16px" name="sym_r_help" class="text-ink-3 q-ml-xs">
					<q-tooltip
						self="top left"
						class="text-body3"
						:offset="[0, 0]"
						style="width: 284px"
					>
						<div>
							{{
								t(
									'The search index excludes any file with a filename or path matching the patterns below. Each line is a separate regular expression.'
								)
							}}
						</div>
					</q-tooltip>
				</q-icon>
			</div>
		</module-title>
		<bt-list v-if="excludedFiles.length > 0">
			<bt-form-item
				v-for="(item, index) in excludedFiles"
				:key="item"
				:title="item"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="index !== excludedFiles.length - 1"
			>
				<q-btn
					class="btn-size-lg"
					icon="sym_r_delete"
					text-color="ink-1"
					no-caps
					@click="deletePattern(item)"
				/>
			</bt-form-item>
		</bt-list>
		<list-empty-component
			v-else
			icon="sym_r_draft"
			:message="
				t('No excluded patterns yet. All files are currently being indexed.')
			"
			:button-label="t('Add pattern')"
			@on-button-click="addPattern"
		/>

		<div
			v-if="excludedFiles.length > 0"
			class="full-width row justify-end q-mt-lg"
		>
			<q-btn
				dense
				class="rebuild-button q-px-md q-py-sm text-body3 text-ink-2"
				:label="t('Add pattern')"
				no-caps
				@click="addPattern"
			/>
		</div>

		<module-title class="q-mt-xl">
			<div class="row justify-start items-center">
				<div>{{ t('Full-Text Search Directories') }}</div>
				<q-icon size="16px" name="sym_r_help" class="text-ink-3 q-ml-xs">
					<q-tooltip
						self="top left"
						class="text-body3"
						:offset="[0, 0]"
						style="width: 284px"
					>
						<div>
							{{
								t(
									'Performs full-text search within all files in the listed directories. Supported formats: pdf, doc/docx, csv, rtf, txt/md/json/xml.'
								)
							}}
						</div>
					</q-tooltip>
				</q-icon>
			</div>
		</module-title>
		<bt-list v-if="fullSearchPaths.length > 0">
			<bt-form-item
				v-for="(item, index) in fullSearchPaths"
				:key="item"
				:title="item"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="index !== fullSearchPaths.length - 1"
			>
				<q-btn
					class="btn-size-lg"
					icon="sym_r_delete"
					text-color="ink-1"
					no-caps
					@click="deleteDirectory(item)"
				/>
			</bt-form-item>
		</bt-list>

		<list-empty-component
			v-else
			icon="sym_r_folder"
			:message="
				t(
					'No custom directories added. Search will only look in default locations.'
				)
			"
			:button-label="t('add_directory')"
			class="q-mb-lg"
			@on-button-click="handleEmptyComponentBtnClick"
		/>

		<div
			v-if="fullSearchPaths.length > 0"
			class="full-width row justify-end q-my-lg"
		>
			<q-btn
				dense
				no-caps
				class="rebuild-button q-px-md q-py-sm text-body3 text-ink-2"
				:label="t('add_directory')"
				@click="handleEmptyComponentBtnClick"
			/>
		</div>
		<transfet-select-to
			ref="transferSelectRef"
			class="q-mt-xs"
			@setSelectPath="setSelectPath"
			:origins="BackupPathOrigins"
			:master-node="true"
		>
			<div />
		</transfet-select-to>
	</bt-scroll-area>
</template>
<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ListEmptyComponent from 'src/components/settings/ListEmptyComponent.vue';
import TransfetSelectTo from 'src/pages/Electron/Transfer/TransfetSelectTo.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import BtFormItem from 'src/components/settings/base/BtFormItem.vue';
import ModuleTitle from 'src/components/settings/ModuleTitle.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import { onMounted, onUnmounted, ref } from 'vue';
import { BackupPathOrigins } from 'src/constant';
import { FilePath } from 'src/stores/files';
import { dataAPIs } from 'src/api/files/v2';
import {
	addExcludePattern,
	addSearchDirectories,
	deleteExcludePattern,
	deleteSearchDirectories,
	getExcludePatterns,
	getSearchDirectories,
	getSearchTaskStatus,
	rebuildSearchTask
} from 'src/api/settings/search';
import { BtDialog } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { decodeUrl } from 'src/utils/encode';

const { t } = useI18n();
const excludedFiles = ref<string[]>([]);
const fullSearchPaths = ref<string[]>([]);
type SearchTaskStatus = 'running' | 'completed';
const taskStatus = ref<SearchTaskStatus>('running');
let taskStatusTimer: NodeJS.Timeout;
const transferSelectRef = ref();

const handleEmptyComponentBtnClick = () => {
	if (transferSelectRef.value) {
		transferSelectRef.value.selectFolder();
	}
};

onMounted(async () => {
	await fetchSearchTaskStatus();
	await fetchExcludedPatterns();
	await fetSearchDirectories();
});

onUnmounted(() => {
	if (taskStatusTimer) {
		clearInterval(taskStatusTimer);
		taskStatusTimer = null;
	}
});

const fetchSearchTaskStatus = async () => {
	taskStatus.value = await getSearchTaskStatus();
	// if (taskStatus.value === 'running') {
	if (taskStatusTimer) {
		clearInterval(taskStatusTimer);
	}
	taskStatusTimer = setInterval(async () => {
		try {
			taskStatus.value = await getSearchTaskStatus();
			if (taskStatus.value === 'completed') {
				clearInterval(taskStatusTimer);
				taskStatusTimer = null;
			}
		} catch (error) {
			console.error('fetchSearchTaskStatus interval', error);
		}
	}, 10 * 1000);
	// }
};

const fetchExcludedPatterns = async () => {
	try {
		excludedFiles.value = await getExcludePatterns();
	} catch (error) {
		console.error('e===>', error);
	}
};

const fetSearchDirectories = async () => {
	try {
		fullSearchPaths.value = await getSearchDirectories();
	} catch (error) {
		console.error('e===>', error);
	}
};

const rebuildTask = async () => {
	try {
		await rebuildSearchTask();
		notifySuccess('success');
		await fetchSearchTaskStatus();
	} catch (error) {
		console.error('e===>', error);
	}
};

const addPattern = () => {
	BtDialog.show({
		title: t('Add pattern'),
		cancel: true,
		prompt: {
			model: '',
			type: 'text', // optional
			name: t('Excluded File Patterns'),
			placeholder: t('Excluded File Patterns')
		},
		okText: t('base.confirm')
	})
		.then((res) => {
			if (res) {
				addExcludePattern([res])
					.then(() => {
						notifySuccess('success');
						fetchExcludedPatterns();
					})
					.catch((error) => {
						console.error('e===>', error);
					});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click ok', err);
		});
};

const deletePattern = (pattern: string) => {
	BtDialog.show({
		title: t('Delete pattern'),
		message: t(
			'Are you sure you want to delete this pattern? The change will take effect after the search index updates.'
		),
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				deleteExcludePattern([pattern])
					.then(() => {
						notifySuccess('success');
						fetchExcludedPatterns();
					})
					.catch((error) => {
						console.error('e===>', error);
					});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};

const deleteDirectory = (directory: string) => {
	BtDialog.show({
		title: t('Delete directory'),
		message: t(
			'Are you sure you want to delete this directory from full-text search? The change will take effect after the search index updates.'
		),
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				deleteSearchDirectories([directory])
					.then(() => {
						notifySuccess('success');
						fetSearchDirectories();
					})
					.catch((error) => {
						console.error('e===>', error);
					});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};

const setSelectPath = async (fileSavePath: FilePath) => {
	console.log(fileSavePath);
	const dataAPI = dataAPIs(fileSavePath.driveType);
	const path = decodeUrl(await dataAPI.formatUploaderPath(fileSavePath.path));
	// console.log(decodeUrl(path));
	if (path) {
		addSearchDirectories([path])
			.then(() => {
				notifySuccess('success');
				fetSearchDirectories();
			})
			.catch((error) => {
				console.error('e===>', error);
			});
	}
};
</script>

<style scoped lang="scss">
.status-button {
	border-radius: 4px;
	padding: 2px 4px;
}

.rebuild-button {
	flex: 0 0 64;
	border: solid 1px $btn-stroke;
}
</style>
