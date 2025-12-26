<template>
	<q-drawer
		show-if-above
		behavior="desktop"
		:width="241"
		class="myDrawer"
		:dark="$q.dark.isActive"
		:class="isFiles ? 'files-border' : ''"
	>
		<BtScrollArea style="height: 100%; width: 100%">
			<bt-menu
				:items="filesStore.menu[origin_id]"
				:modelValue="filesStore.activeMenu(origin_id).id"
				:sameActiveable="false"
				@select="selectHandler"
				style="width: 239px"
				class="title-norla"
				active-class="text-subtitle2 bg-yellow-soft text-ink-1"
			>
				<template #extra-MyLibraries> </template>

				<template #extra-Sync>
					<div class="text-subtitle1 q-pr-none row item-scenter justify-end">
						<div v-if="$q.platform.is.electron && menuStore.reposHasSync">
							<q-btn
								v-if="menuStore.syncStatus"
								class="btn-size-xs btn-no-text btn-no-border text-ink-2"
								icon="sym_r_pause_circle"
								text-color="ink-2"
								@click="menuStore.updateSyncStatus"
							>
								<q-tooltip> {{ t('files.click_to_pause') }}</q-tooltip>
							</q-btn>

							<q-btn
								v-if="!menuStore.syncStatus"
								class="btn-size-xs btn-no-text btn-no-border text-ink-2"
								icon="sym_r_autoplay"
								text-color="ink-2"
								@click="menuStore.updateSyncStatus"
							>
								<q-tooltip> {{ t('files.click_to_continue') }}</q-tooltip>
							</q-btn>
						</div>
						<q-btn
							class="btn-size-xs btn-no-text btn-no-border text-ink-1"
							icon="sym_r_add_circle"
							text-color="ink-2"
							@click="handleNewLib($event)"
						>
							<q-tooltip> {{ t('files.new_library') }}</q-tooltip>
						</q-btn>
					</div>
				</template>

				<template
					v-slot:[`icon-${menu.id}`]
					v-for="menu in filesStore.menu[origin_id][1]?.children"
					:key="menu.id"
				>
					<q-icon class="item-icon" rounded :name="menu.icon" size="24px">
						<q-circular-progress
							v-if="
								$q.platform.is.electron &&
								syncStatusInfo[getSyncStatus(menu.id)] &&
								getSyncStatus(menu.id) == SYNC_STATE.ING &&
								menuStore.syncReposLastStatusMap[menu.id].percent &&
								menuStore.syncReposLastStatusMap[menu.id].percent > 0
							"
							rounded
							:value="menuStore.syncReposLastStatusMap[menu.id].percent"
							size="12px"
							:thickness="0.4"
							color="light-blue-default"
							track-color="light-blue-alpha"
							class="sync-icon bg-background-1"
						/>

						<q-icon
							v-else-if="
								$q.platform.is.electron &&
								syncStatusInfo[getSyncStatus(menu.id)]
							"
							:name="syncStatusInfo[getSyncStatus(menu.id)].icon"
							size="12px"
							color="white"
							class="sync-icon"
							:style="{
								background: syncStatusInfo[getSyncStatus(menu.id)].color
							}"
						>
						</q-icon>
					</q-icon>
				</template>

				<template
					v-slot:[`extra-${menu.id}`]
					v-for="menu in filesStore.menu[origin_id][1]?.children"
					:key="menu.id"
				>
					<q-btn
						class="btn-size-xs btn-no-text btn-no-border text-ink-1"
						icon="more_horiz"
						text-color="ink-2"
					>
						<q-tooltip>{{ t('files.operate') }}</q-tooltip>
						<PopupMenu
							:item="{
								...menu,
								isDir: true
							}"
							from="sync"
							:isSide="true"
						/>
					</q-btn>
				</template>
			</bt-menu>
		</BtScrollArea>
	</q-drawer>
</template>

<script setup lang="ts">
import { useQuasar } from 'quasar';
import { onMounted, defineProps } from 'vue';
import { useRoute } from 'vue-router';
import { useDataStore } from '../../stores/data';
import { syncStatusInfo, useMenuStore } from '../../stores/files-menu';
import { useOperateinStore } from './../../stores/operation';
import PopupMenu from '../../components/files/popup/PopupMenu.vue';
import { OPERATE_ACTION, SYNC_STATE } from '../../utils/contact';
import { useI18n } from 'vue-i18n';
import { useFilesStore, FilesIdType } from './../../stores/files';
import { DriveType } from '../../utils/interface/files';

const $q = useQuasar();
const Route = useRoute();
const store = useDataStore();
const menuStore = useMenuStore();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();
const isFiles = process.env.APPLICATION == 'FILES';

const { t } = useI18n();

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

onMounted(async () => {
	await filesStore.getMenu();
});

const selectHandler = async (value: any) => {
	const path = await filesStore.formatRepotoPath(value.item);
	filesStore.setBrowserUrl(path, value.item.driveType, true, props.origin_id);
	filesStore.resetSelected();
};

const handleNewLib = (e: any) => {
	operateinStore.handleFileOperate(
		props.origin_id,
		e,
		Route,
		OPERATE_ACTION.CREATE_REPO,
		DriveType.Sync,
		// eslint-disable-next-line @typescript-eslint/no-unused-vars
		async (_action: OPERATE_ACTION, _data: any) => {
			store.closeHovers();
		}
	);
};

const getSyncStatus = (repo_id: string) => {
	const status = menuStore.syncReposLastStatusMap[repo_id]
		? menuStore.syncReposLastStatusMap[repo_id].status
		: 0;

	if (status > 0) {
		if (!menuStore.syncStatus) {
			return -1;
		}
	}
	return status;
};
</script>

<style lang="scss">
.myDrawer {
	overflow: hidden;
	padding-top: 6px;
}

.files-border {
	border-right: 1px solid $separator;
}

.sync-icon {
	position: absolute;
	left: -1.5px;
	top: 12px;
	cursor: pointer;
	border-radius: 12px;
	font-variation-settings: 'FILL' 1, 'wght' 300, 'GRAD' 0, 'opsz' 20;
}
</style>
