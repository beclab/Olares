<template>
	<page-title-component :show-back="true" :title="t('File Search')" />
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature
			image="settings/imgs/root/search.svg"
			:label="t('Search Index')"
			:description="t('Rebuilds search index using the latest settings.')"
			:button="t('Rebuild')"
			@on-button-click="rebuildTask"
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
		</app-menu-feature>

		<div
			v-if="showFailList"
			class="text-body3 text-ink-3 row justify-start items-center"
		>
			<q-icon class="q-ml-sm" name="sym_r_report" size="16px" />
			<span>{{ t('Failed to create full-text index for some files.') }}</span>
			<span class="text-info cursor-pointer" @click="viewFailedFileList">
				{{ t('Click to view the list') }}
			</span>
		</div>

		<module-title class="q-mt-lg">
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
				:key="item.pattern + '-' + index"
				:margin-top="false"
				:chevron-right="false"
				:widthSeparator="index !== excludedFiles.length - 1"
			>
				<template #title>
					<excluded-pattern-title :text="item.pattern" />
				</template>
				<q-btn
					v-if="!item.must"
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
			:origins="SearchPathOrigins"
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
import ExcludedPatternTitle from 'src/pages/settings/Search/ExcludedPatternTitle.vue';
import { notifySuccess, notifyFailed } from 'src/utils/settings/btNotify';
import { onMounted, onUnmounted, ref } from 'vue';
import { ExcludePatternItem, SearchPathOrigins } from 'src/constant';
import { FilePath } from 'src/stores/files';
import { dataAPIs } from 'src/api/files/v2';
import { decodeUrl } from 'src/utils/encode';
import { BtDialog } from '@bytetrade/ui';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
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

const MAX_EXCLUDED_PATTERN_LENGTH = 1024;

const showFailList = ref(false);
const { t } = useI18n();
const router = useRouter();
const excludedFiles = ref<ExcludePatternItem[]>([]);
const fullSearchPaths = ref<string[]>([]);
type SearchTaskStatus = 'running' | 'completed';
const taskStatus = ref<SearchTaskStatus>('running');
let taskStatusTimer: ReturnType<typeof setInterval> | null = null;
const transferSelectRef = ref();

const handleEmptyComponentBtnClick = () => {
	if (transferSelectRef.value) {
		transferSelectRef.value.selectFolder();
	}
};

const viewFailedFileList = () => {
	router.push('/search/failed-files');
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
	if (taskStatusTimer) {
		clearInterval(taskStatusTimer);
		taskStatusTimer = null;
	}

	const applyMergedPayload = (raw: unknown) => {
		const row = (raw as { data?: Record<string, unknown> })?.data ?? raw;
		const { status, full_content_task_error } = row as {
			status: SearchTaskStatus;
			full_content_task_error?: boolean;
		};
		taskStatus.value = status;
		showFailList.value = Boolean(full_content_task_error);
	};

	try {
		const res = await getSearchTaskStatus();
		applyMergedPayload(res);
	} catch (error) {
		console.error('fetchSearchTaskStatus', error);
		return;
	}

	if (taskStatus.value !== 'running') {
		return;
	}

	taskStatusTimer = setInterval(async () => {
		try {
			const res = await getSearchTaskStatus();
			applyMergedPayload(res);
			if (taskStatus.value === 'completed') {
				clearInterval(taskStatusTimer!);
				taskStatusTimer = null;
			}
		} catch (error) {
			console.error('fetchSearchTaskStatus interval', error);
		}
	}, 10 * 1000);
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
		const list = (await getSearchDirectories()) as unknown as string[];
		fullSearchPaths.value = list;
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
			type: 'text',
			name: t('Excluded file patterns'),
			placeholder: t('Excluded file patterns'),
			isValid: (value: string) => {
				const trimmed = String(value ?? '').trim();
				if (!trimmed) {
					notifyFailed(t('Excluded pattern cannot be empty or only spaces'));
					return false;
				}
				if (trimmed.length > MAX_EXCLUDED_PATTERN_LENGTH) {
					notifyFailed(
						t('Excluded pattern is too long (max {max} characters)', {
							max: MAX_EXCLUDED_PATTERN_LENGTH
						})
					);
					return false;
				}
				return true;
			}
		},
		okText: t('base.confirm')
	})
		.then((res) => {
			if (!res) {
				return;
			}
			const trimmed = String(res).trim();
			addExcludePattern([trimmed])
				.then(async () => {
					notifySuccess('success');
					await fetchExcludedPatterns();
					await fetchSearchTaskStatus();
				})
				.catch((error) => {
					console.error('e===>', error);
				});
		})
		.catch((err) => {
			console.log('click error', err);
		});
};

const deletePattern = (item: ExcludePatternItem) => {
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
				deleteExcludePattern([item.pattern])
					.then(async () => {
						notifySuccess('success');
						await fetchExcludedPatterns();
						await fetchSearchTaskStatus();
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
					.then(async () => {
						notifySuccess('success');
						await fetSearchDirectories();
						await fetchSearchTaskStatus();
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
			.then(async () => {
				notifySuccess('success');
				await fetSearchDirectories();
				await fetchSearchTaskStatus();
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
