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
				active-class="text-subtitle2 bg-menu-active-color"
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
import { onMounted, defineProps, onUnmounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useDataStore } from '../../stores/data';
import { syncStatusInfo, useMenuStore } from '../../stores/files-menu';
import { useOperateinStore } from './../../stores/operation';
import PopupMenu from '../../components/files/popup/PopupMenu.vue';
import { OPERATE_ACTION, SYNC_STATE } from '../../utils/contact';
import { useI18n } from 'vue-i18n';
import { useFilesStore, FilesIdType } from './../../stores/files';
import { DriveType } from '../../utils/interface/files';
import { busOff, busOn } from 'src/utils/bus';
import { useTermipassStore } from '../../stores/termipass';
import { UserStatusActive } from '../../utils/checkTerminusState';

const $q = useQuasar();
const Route = useRoute();
const store = useDataStore();
const menuStore = useMenuStore();
const operateinStore = useOperateinStore();
const filesStore = useFilesStore();
const termipassStore = useTermipassStore();
const isFiles = process.env.APPLICATION == 'FILES';

const { t } = useI18n();

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

// The sync / network-drive lists are built from API responses in
// `filesStore.getMenu()`. If that fetch fails (e.g. transient network issue),
// nothing re-runs it, leaving the drawer empty until a manual reload. Mirroring
// FilesPage.vue: when the overall user/olares status recovers from an error,
// refetch the menu so those API-driven lists repopulate automatically.
const menuError = ref<any>(null);
let menuRetryTimer: ReturnType<typeof setTimeout> | null = null;

const clearMenuRetryTimer = () => {
	if (menuRetryTimer) {
		clearTimeout(menuRetryTimer);
		menuRetryTimer = null;
	}
};

const checkMenuStatus = (status: any) => {
	if (status === UserStatusActive.error) {
		// Re-entered an error before a pending recovery fired: drop the stale
		// retry so it can't run against the now-unhealthy session.
		clearMenuRetryTimer();
		menuError.value = status;
		return;
	}
	// No active session (logged out / not initialized): `totalStatus` is
	// undefined, so `isError` is undefined too. This is NOT a recovery — cancel
	// any pending refetch and drop the stale error marker, since getMenu() would
	// run without a session.
	if (status === undefined || status === null) {
		clearMenuRetryTimer();
		menuError.value = undefined;
		return;
	}
	if (menuError.value != null && menuError.value != undefined) {
		clearMenuRetryTimer();
		menuRetryTimer = setTimeout(
			async () => {
				menuRetryTimer = null;
				// Guard against logout between scheduling and firing: only refetch
				// while there is still an active session.
				if (!termipassStore.totalStatus) {
					return;
				}
				try {
					await filesStore.getMenu();
					// Clear the error marker only after a successful refetch. If
					// getMenu() fails while the overall status stays healthy (no new
					// error→recovery transition follows), keeping `menuError` set lets
					// a later recovery reschedule the retry instead of leaving the
					// sync / network-drive lists empty until a full reload.
					menuError.value = undefined;
				} catch (e) {
					// keep menuError set so a later recovery retries
				}
			},
			termipassStore.isLocal ? 5000 : 2000
		);
	}
};

watch(
	() => termipassStore.totalStatus?.isError,
	(newVal) => {
		checkMenuStatus(newVal);
	},
	// `immediate` so a drawer that mounts while already in the error state still
	// records `menuError`; otherwise the recovery refetch is skipped and the
	// sync / network-drive lists stay empty after connectivity returns.
	{ immediate: true }
);

onMounted(async () => {
	busOn('reposUpdate', reposUpdate);
	await filesStore.getMenu();
});

onUnmounted(() => {
	busOff('reposUpdate', reposUpdate);
	clearMenuRetryTimer();
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

const reposUpdate = async (isRemove?: boolean) => {
	await filesStore.getMenu();
	if (isRemove == true) {
		await filesStore.openFirstMenuRepo(props.origin_id);
	}
};
</script>

<style lang="scss">
.myDrawer {
	overflow: hidden;
	padding-top: 6px;
}

.myDrawer .bg-menu-active-color {
	background: $theme-menu-bg-hover;
	color: $theme-menu-color-hover;
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
