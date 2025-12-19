<template>
	<page-title-component :show-back="false" :title="t('knowledge_bases')">
		<template v-slot:end>
			<div
				class="add-btn row justify-center items-center"
				@click="addOrEditSearchFolderPath()"
			>
				<q-icon size="16px" name="sym_r_add" color="ink-1" />
				<div class="text-body3 add-title">
					{{ t('create_new') }}
				</div>
			</div>
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<empty-component
			v-if="fileStore.datasets.length === 0"
			:empty-item="t('knowledge')"
		/>
		<div
			class="q-list-class"
			v-for="folder in fileStore.datasets"
			:key="folder.datasetID"
		>
			<div class="text-subtitle1 text-ink-1 q-mb-md">
				{{ folder.datasetName }}
			</div>
			<bt-separator />
			<div class="row items-center justify-between folder-content">
				<div class="row path_content">
					<q-img
						class="directory-image q-mr-sm"
						src="settings/folder_documents.svg"
						noSpinner
					/>
					<div class="text-body2 text-ink-1 path">
						{{ folder.paths != null ? folder.paths.join(',') : '' }}
					</div>
				</div>
				<q-btn dense flat icon="sym_r_more_horiz">
					<q-menu :offset="[90, 5]">
						<q-list dense class="popup-list">
							<q-item
								class="row items-center justify-start popup-item text-ink-2"
								style="padding: 8px; border-radius: 4px"
								clickable
								v-close-popup
								@click="addOrEditSearchFolderPath(folder)"
							>
								<q-icon name="sym_r_edit_square" size="20px" class="q-mr-sm" />
								<q-item-section class="text-body3 text-ink-2">{{
									t('edit')
								}}</q-item-section>
							</q-item>
							<q-item
								class="row items-center justify-start popup-item text-ink-2"
								style="padding: 8px; border-radius: 4px"
								clickable
								:disable="folder.default"
								v-close-popup
								@click="showRemoveDialog(folder)"
							>
								<q-icon name="sym_r_delete" size="20px" class="q-mr-sm" />
								<q-item-section class="text-body3 text-ink-2">{{
									t('remove')
								}}</q-item-section>
							</q-item>
						</q-list>
					</q-menu>
				</q-btn>
			</div>
			<div class="row items-center text-ink-2 text-body3">
				<div class="row items-center">
					<q-icon name="sym_r_description" size="16px" class="q-mr-xs" />
					<div>{{ (folder.indexDocNum || 0) + ' ' + t('docs') }}</div>
				</div>
				<div class="row items-center q-ml-lg">
					<q-icon name="sym_r_robot_2" size="16px" class="q-mr-xs" />
					<div>
						{{
							t('number_linked_agent', {
								number: folder.linkedAgentNum || 0
							})
						}}
					</div>
				</div>
			</div>
		</div>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { useFilesStore, DatasetFolder } from 'src/stores/settings/files';
import AddKnowledgeBaseDialog from './dialog/AddKnowledgeBaseDialog.vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ReminderDialogComponent from 'src/components/settings/ReminderDialogComponent.vue';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';

const { t } = useI18n();

const $q = useQuasar();
const fileStore = useFilesStore();

onMounted(() => {
	fileStore.GetDatasetFolderStatus();
	// fileStore.UpdateDatasetFolderPaths('258d7bf5-e3e0-4f67-acfb-5dbcfd28b201', [
	// 	'/data/Home/Documents'
	// ]);
});

const addOrEditSearchFolderPath = (folder?: DatasetFolder) => {
	$q.dialog({
		component: AddKnowledgeBaseDialog,
		componentProps: {
			datasetID: folder ? folder.datasetID : ''
		}
	}).onOk(async (data: { name: string; paths: string[] }) => {
		// if (folder) {

		// }
		$q.loading.show();
		try {
			await fileStore.UpdateDatasetFolderPaths(
				folder ? folder.datasetID : undefined,
				folder ? undefined : data.name,
				data.paths,
				folder ? 0 : 1
			);
			notifySuccess();
		} catch (error) {
			console.log(error);
			notifyFailed(error.message);
		} finally {
			$q.loading.hide();
		}
	});
};

const showRemoveDialog = (folder: DatasetFolder) => {
	$q.dialog({
		component: ReminderDialogComponent,
		componentProps: {
			title: t('remove_knowledge_base', {
				base: folder.datasetName
			}),
			message: t('remove_knowledge_base_message', {
				base: `'${folder.datasetName}'`
			}),
			useCancel: true,
			isReminder: true,
			confirmText: t('remove')
		}
	}).onOk(async () => {
		$q.loading.show();
		try {
			await fileStore.UpdateDatasetFolderPaths(
				folder.datasetID,
				undefined,
				undefined,
				-1
			);
		} catch (error) {
			notifyFailed(error.message);
		} finally {
			$q.loading.hide();
		}
	});
};
</script>

<style scoped lang="scss">
.directory-tips {
	margin-top: 12px;
	margin-bottom: 12px;
	color: $ink-2;
}

.add-btn {
	border-radius: 8px;
	padding: 6px 12px;
	border: 1px solid $separator;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}
.add-btn:hover {
	background-color: $background-3;
}
.directory-image {
	width: 24px;
	height: 24px;
}

.folder-content {
	width: 100%;
	// height: 56px;
	// background-color: red;
	padding-top: 12px;

	.path_content {
		width: calc(100% - 40px);
		word-wrap: word-break;
		word-break: break-all;

		.path {
			width: calc(100% - 40px);
			margin-top: 2px;
			margin-bottom: 10px;
			// background: green;
		}
	}
}

.popup-list {
	width: 120px;
	padding: 8px;
	border-radius: 8px;
	box-shadow: 0px 4px 10px 0px #00000033;
}
.empty-image {
	margin-top: 144px;
	width: 160px;
	height: 160px;
}
</style>
