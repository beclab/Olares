<template>
	<div class="desktop-box">
		<div class="desktop_app_box">
			<div
				class="desktop_app"
				:style="`height:${DOCKER_APP_TOTAL_HEIGHT + 180}px;`"
			>
				<div class="desktop_user_avatar">
					<div class="desktop_avatar" @click="openProfile">
						<TerminusAvatar :info="tokenStore.terminus" :size="36" />
					</div>

					<div class="desktop_avatar_border"></div>
				</div>
				<div
					class="column_apps_all"
					id="app_dock_list"
					@dragover.prevent="onDragOver"
					@drop.prevent="onDockDrop"
					@dragleave="onDockDragLeave"
				>
					<template v-for="element in appStore.dockerApps" :key="element.id">
						<img
							ref="imgRef"
							@click="openWindow(element)"
							:id="element.id"
							:style="`position:absolute;top:${element.top}px;width:${
								appStore.DOCKER_APP_SIZE
							}px;height:${appStore.DOCKER_APP_SIZE}px;left:${
								(60 - appStore.DOCKER_APP_SIZE) / 2
							}px;border-radius: ${borderRadiusFormat(
								appStore.DOCKER_APP_SIZE,
								appStore.DOCKER_APP_SIZE
							)}px;transition:${
								transitionsEnabled
									? 'top ' + DOCK_REORDER_MS + 'ms ease'
									: 'none'
							}`"
							draggable="true"
							@dragstart="onDragStart"
							@dragenter.prevent="onDragOver"
							@dragover.prevent="onDragOver"
							@dragend="onDragEnd"
							:src="element.icon"
							@contextmenu.prevent="onContextmenu(element)"
						/>
						<div
							v-if="isBookmarkDockItem(element)"
							class="dock-bookmark-badge"
							:style="`position:absolute;top:${element.top}px;left:${
								(60 - appStore.DOCKER_APP_SIZE) / 2
							}px;width:${appStore.DOCKER_APP_SIZE}px;height:${
								appStore.DOCKER_APP_SIZE
							}px;pointer-events:none;`"
						>
							<BookmarkBadge />
						</div>
						<div
							:class="{
								img_parent: element.show_dot && !isAppMinimized(element.id),
								img_parent_minimized: isAppMinimized(element.id)
							}"
							:style="`position:absolute;top:${
								element.top + 17
							}px;right: -3px;`"
						></div>
						<div
							:class="contextmenuId == element.id ? 'menu_open' : 'menu_hide'"
							:style="`position:absolute;top:${element.top}px;left:70px;`"
						>
							<div
								class="item"
								v-if="!isAppOpen(element.id)"
								@click="openWindow(element)"
							>
								{{ t('dock_operation.open') }}
							</div>

							<div
								class="item"
								v-if="isAppOpen(element.id) && isAppMinimized(element.id)"
								@click="onShow(element)"
							>
								{{ t('dock_operation.show') }}
							</div>

							<div
								class="item"
								v-if="isAppOpen(element.id) && !isAppMinimized(element.id)"
								@click="mini(element)"
							>
								{{ t('dock_operation.hide') }}
							</div>

							<div
								class="item"
								v-if="element.show_dot"
								@click="updateAppTemp(element)"
							>
								<span>{{ t('dock_operation.keepInDock') }}</span>
								<span class="q-ml-md" style="width: 30px"
									><q-icon
										v-if="!element.is_temp"
										class="icon"
										name="check"
										size="12px"
								/></span>
							</div>

							<q-separator class="q-my-xs q-mx-sm" color="#d9d9d9" />

							<div
								class="item red_font"
								v-if="!element.is_temp && !element.show_dot"
								@click="removeAppInDock(element.id)"
							>
								{{ t('dock_operation.removeFromDock') }}
							</div>

							<div
								class="item red_font"
								v-if="element.show_dot"
								@click="onClose(element)"
							>
								{{ t('dock_operation.quit') }}
							</div>
							<div class="context_triangle"></div>
						</div>
					</template>
				</div>

				<div class="desktop_message desktop_user_avatar">
					<div class="desktop_avatar_border"></div>
					<div class="dock_icon" @click="toSearch">
						<img src="../../assets/desktop/dock-search.svg" alt="search" />
					</div>

					<div class="dock_icon" @click="showNotification">
						<img src="../../assets/desktop/dock-notify.svg" alt="notify" />
						<div class="badge" v-if="notificationStore.newMessage" />
					</div>
					<!-- <div class="dock_icon" @click="onLogout">
						<img src="../../assets/desktop/dock-logout.svg" alt="logout" />
					</div> -->
				</div>
			</div>
		</div>
	</div>
</template>
<script lang="ts" setup>
import {
	ref,
	computed,
	watch,
	nextTick,
	onMounted,
	onUnmounted,
	PropType
} from 'vue';
import { Loading, useQuasar } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { DockerAppInfo, AppClickInfo, WindowInfo } from './type/types';
import { useTokenStore } from '../../stores/desktop/token';
import { useApplicationStore } from '../../stores/desktop/app';
import { borderRadiusFormat } from '../../utils/desktop/utils';
import ConfirmDialog from './components/ConfirmDialog.vue';
import BookmarkBadge from './components/BookmarkBadge.vue';
import { useI18n } from 'vue-i18n';

import { useNotificationStore } from '../../stores/desktop/notification';
import {
	DockItem,
	DockLayout,
	LAYOUT_SCHEMA_VERSION,
	rawIdToKey,
	isBookmarkRawId
} from 'src/types/desktop-layout';
import { ENTRANCE_STATUS, APP_STATUS } from 'src/constant/constants';
import { isDoingState } from 'src/constant/config';

function pushDockLayoutPC() {
	// Writes the single shared dock layout (mobile dock is read-only and
	// follows this same synced data).
	const items: DockItem[] = [];
	const seen = new Set<string>();
	let slot = 0;
	const sorted = [...appStore.dockerApps].sort((a, b) => a.top - b.top);
	for (const a of sorted) {
		if (a.is_temp) continue;
		const raw = a.id;
		const rawId = raw.startsWith('bdock:') ? raw.slice(6) : raw;
		// launchpad is now a normal draggable + persisted dock item; keep it in
		// the saved layout so its slot survives reorders (was previously stripped
		// and force-reinjected at the front on every projection).
		if (!rawId || seen.has(rawId)) continue;
		seen.add(rawId);
		items.push({ key: rawIdToKey(rawId), slot: slot++ });
	}
	const prev = appStore.layoutBySurface.dock;
	const next: DockLayout = {
		schemaVersion: LAYOUT_SCHEMA_VERSION,
		items,
		capacity: prev?.capacity,
		version: prev?.version ?? 0,
		updatedAt: Date.now(),
		updatedBy: ''
	};
	appStore.saveDockLayoutDebounced(next);
}

const props = defineProps({
	drag_launch_app: {
		type: String,
		required: true
	},
	window_infos: {
		type: Array as PropType<WindowInfo[]>,
		required: false,
		default: () => []
	}
});

const emits = defineEmits([
	'changeSearchDialog',
	'appClick',
	'onClose',
	'mini'
]);

const tokenStore = useTokenStore();
const appStore = useApplicationStore();
const notificationStore = useNotificationStore();

const { t } = useI18n();
const contextmenuId = ref<string>('');
const $q = useQuasar();
const DOCK_REORDER_MS = 180;

type DockDragKind = 'none' | 'reorder' | 'fromLaunchpad';
let dragKind: DockDragKind = 'none';
let origin_app: DockerAppInfo | null = null;
let choose = -1;
let pendingInsertIndex = -1;
let reorderSnapshot: Map<string, number> | null = null;
let dropHandled = false;

const transitionsEnabled = ref(false);
const dragPreviewActive = ref(false);

const onLogout = async () => {
	$q.dialog({
		component: ConfirmDialog,
		componentProps: {
			title: t('Log out'),
			message: t('Your changes may not be saved.'),
			showCancel: true
		}
	}).onOk(async () => {
		try {
			Loading.show();
			await tokenStore.logout();
			const auth_url = tokenStore.getAuthURL() + '?logout=1';
			window.location.replace(auth_url);
		} catch (err) {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: (err as Error).message
			});
		} finally {
			Loading.hide();
		}
	});
};

const DOCKER_APP_TOTAL_HEIGHT = computed(() => {
	const n = appStore.dockerApps.length + (dragPreviewActive.value ? 1 : 0);
	if (n <= 0) {
		return appStore.DOCKER_APP_START_GAP + appStore.DOCKER_APP_END_GAP;
	}
	return (
		appStore.DOCKER_APP_START_GAP +
		n * appStore.DOCKER_APP_SIZE +
		(n - 1) * appStore.DOCKER_APP_GAP +
		appStore.DOCKER_APP_END_GAP
	);
});

const openAppIds = computed(() => {
	const result = new Set<string>();
	for (const info of props.window_infos) {
		result.add('bdock:' + info.id);
	}
	return result;
});

const minimizedAppIds = computed(() => {
	const result = new Set<string>();
	for (const info of props.window_infos) {
		if (!info.is_show) {
			result.add('bdock:' + info.id);
		}
	}
	return result;
});

const isAppOpen = (id: string) => openAppIds.value.has(id);
const isAppMinimized = (id: string) => minimizedAppIds.value.has(id);

const isBookmarkDockItem = (item: DockerAppInfo) =>
	item.kind === 'bookmark' || isBookmarkRawId(appStore.normalizeAppId(item.id));

const openWindow = async (item: DockerAppInfo) => {
	// A dock entry only stores layout data, not the live app state. Resolve the
	// backing app so a paused / crashed / installing app opened from the dock
	// shows the same dialog the launchpad shows, instead of opening a window
	// that fails to load.
	const app = appStore.resolveAppForOpen(item.id);
	if (app) {
		if (isDoingState(app.fatherState)) {
			return;
		}
		if (
			app.state === ENTRANCE_STATUS.STOPPED ||
			app.fatherState === APP_STATUS.UPGRADE.FAILED
		) {
			$q.dialog({
				component: ConfirmDialog,
				componentProps: {
					title: t('confirmation'),
					message: t('message_desktop.suspended'),
					icon: item.icon,
					okTitle: t('i_got_it')
				}
			});
			return;
		}
		if (app.state === ENTRANCE_STATUS.NOT_READY) {
			$q.dialog({
				component: ConfirmDialog,
				componentProps: {
					title: t('confirmation'),
					message: t('message_desktop.crashed'),
					icon: item.icon,
					okTitle: t('i_got_it')
				}
			});
			return;
		}
	}

	emits('appClick', {
		appid: item.id,
		data: {}
	} as AppClickInfo);
};

const onShow = (item: DockerAppInfo) => {
	openWindow(item);
};

const mini = (item: DockerAppInfo) => {
	const curItem = JSON.parse(JSON.stringify(item));
	if (curItem.id.startsWith('bdock:')) {
		curItem.id = curItem.id.substring(6);
	}
	emits('mini', curItem);
};

const onClose = (item: DockerAppInfo) => {
	const curItem = JSON.parse(JSON.stringify(item));
	if (curItem.is_temp) {
		appStore.remove_app_on_docker(curItem.id, true);
	}
	if (curItem.id.startsWith('bdock:')) {
		curItem.id = curItem.id.substring(6);
	}
	emits('onClose', curItem);
};

const openProfile = async () => {
	emits('appClick', {
		appid: appStore.profile_id,
		data: {}
	} as AppClickInfo);
};

const onContextmenu = (element: DockerAppInfo) => {
	if (element.id === 'bdock:launchpad') {
		return false;
	}
	contextmenuId.value = element.id;
};

const getAppIndexInDock = (id: string) =>
	appStore.dockerApps.findIndex((app) => app.id === id);

const normalizeDockId = (appid: string) =>
	appid.startsWith('bdock:') ? appid : 'bdock:' + appid;

const topForSlot = (slot: number) =>
	appStore.DOCKER_APP_START_GAP +
	(appStore.DOCKER_APP_SIZE + appStore.DOCKER_APP_GAP) * slot;

const visualOrder = (excludeId?: string) =>
	appStore.dockerApps
		.filter((a) => !excludeId || a.id !== excludeId)
		.slice()
		.sort((a, b) => a.top - b.top);

const setTop = (id: string, top: number) => {
	const item = appStore.dockerApps.find((a) => a.id === id);
	if (item) item.top = top;
};

const computeInsertIndex = (
	e: any,
	excludeId?: string
): { index: number; visual: DockerAppInfo[] } => {
	const visual = visualOrder(excludeId);
	const el = e.target as HTMLElement | null;
	if (el && el.id && el.id.startsWith('bdock:') && el.id !== excludeId) {
		const pos = visual.findIndex((a) => a.id === el.id);
		if (pos >= 0) {
			const rect = el.getBoundingClientRect();
			const before = e.clientY < rect.top + rect.height / 2;
			return { index: before ? pos : pos + 1, visual };
		}
	}
	const listEl = document.getElementById('app_dock_list');
	if (listEl) {
		const rect = listEl.getBoundingClientRect();
		const pitch = appStore.DOCKER_APP_SIZE + appStore.DOCKER_APP_GAP;
		const idx = Math.round(
			(e.clientY - rect.top - appStore.DOCKER_APP_START_GAP) / pitch
		);
		return { index: Math.max(0, Math.min(visual.length, idx)), visual };
	}
	return { index: visual.length, visual };
};

const reflowDockNormal = () => {
	const visual = visualOrder();
	for (let i = 0; i < visual.length; i++) {
		setTop(visual[i].id, topForSlot(i));
	}
};

const resetDragState = () => {
	if (dragKind === 'fromLaunchpad') {
		reflowDockNormal();
	}
	for (const a of appStore.dockerApps) {
		document.getElementById(a.id)?.classList.remove('drag-start');
	}
	dragKind = 'none';
	origin_app = null;
	choose = -1;
	pendingInsertIndex = -1;
	dropHandled = false;
	dragPreviewActive.value = false;
	reorderSnapshot = null;
	appStore.setLayoutDragLock(false);
};

const onDragStart = (e: any) => {
	resetDragState();

	const idx = getAppIndexInDock(e.target.id);
	if (idx < 0) return;

	dragKind = 'reorder';
	origin_app = appStore.dockerApps[idx];
	choose = idx;
	reorderSnapshot = new Map(appStore.dockerApps.map((a) => [a.id, a.top]));
	appStore.setLayoutDragLock(true);

	e.dataTransfer.setData('text', e.target.id);
	e.dataTransfer.dropEffect = 'move';
	e.dataTransfer.effectAllowed = 'move';
	e.target.classList.add('drag-start');
};

const applyReorder = (e: any) => {
	const origin = origin_app;
	if (!origin) return;
	const { index, visual } = computeInsertIndex(e, origin.id);
	const ordered = [...visual];
	ordered.splice(index, 0, origin);
	for (let i = 0; i < ordered.length; i++) {
		setTop(ordered[i].id, topForSlot(i));
	}
};

const applyLaunchpadPreview = (e: any) => {
	const { index, visual } = computeInsertIndex(e);
	pendingInsertIndex = index;
	dragPreviewActive.value = true;
	for (let i = 0; i < visual.length; i++) {
		setTop(visual[i].id, topForSlot(i < index ? i : i + 1));
	}
};

const ensureDragKind = () => {
	if (dragKind !== 'none') return;
	if (origin_app) {
		dragKind = 'reorder';
	} else if (props.drag_launch_app) {
		dragKind = 'fromLaunchpad';
		appStore.setLayoutDragLock(true);
	}
};

const onDragOver = (e: any) => {
	e.preventDefault();
	ensureDragKind();
	if (dragKind === 'reorder') {
		applyReorder(e);
	} else if (dragKind === 'fromLaunchpad') {
		applyLaunchpadPreview(e);
	}
};

const readDragText = (e: any): string => {
	try {
		return (e && e.dataTransfer && e.dataTransfer.getData('text')) || '';
	} catch {
		return '';
	}
};

const commitLaunchpadDrop = (rawArg?: string) => {
	const raw = rawArg || props.drag_launch_app || '';
	const rawId =
		raw.startsWith('bdesk:') || raw.startsWith('bdock:')
			? raw.substring(6)
			: raw;
	if (!rawId) return;

	const existing = appStore.find_app('bdock:' + rawId);
	if (existing) {
		// Already present (typically a temp icon for an app that's currently
		// open): promote it to a pinned item so it survives the app closing,
		// instead of silently doing nothing and letting it vanish later.
		if (existing.is_temp) {
			existing.is_temp = false;
			// Move it to where the user actually dropped it. Without this the
			// promoted icon keeps the temp slot's `top`, so pushDockLayoutPC()
			// (which orders by `top`) would persist it at the wrong position
			// and ignore pendingInsertIndex entirely.
			const order = visualOrder(existing.id);
			const tempPos = visualOrder().findIndex((a) => a.id === existing.id);
			let target = pendingInsertIndex >= 0 ? pendingInsertIndex : order.length;
			// pendingInsertIndex was measured against the full visual order
			// (temp item included); dropping the item out shifts every slot
			// after it left by one, so compensate before splicing.
			if (tempPos >= 0 && target > tempPos) target -= 1;
			target = Math.max(0, Math.min(target, order.length));
			order.splice(target, 0, existing);
			for (let i = 0; i < order.length; i++) {
				setTop(order[i].id, topForSlot(i));
			}
			appStore.updateAppInDock(existing);
			pushDockLayoutPC();
		}
		return;
	}

	const index =
		pendingInsertIndex >= 0 ? pendingInsertIndex : appStore.dockerApps.length;
	appStore.add_app_on_docker_at(rawId, index, false, false);
	pushDockLayoutPC();
};

const persistReorder = () => {
	reflowDockNormal();
	pushDockLayoutPC();
};

const restoreReorder = () => {
	if (!reorderSnapshot) return;
	for (const [id, top] of reorderSnapshot) {
		setTop(id, top);
	}
	// We reverted to the pre-drag local layout without persisting, so the live
	// dock is now authoritative. Drop any remote dock projection stashed during
	// the drag; otherwise unlocking layout sync in resetDragState() would replay
	// it and clobber the restore the user just got. (Launch layout is unrelated
	// to a dock drag, so it still replays normally on unlock.)
	appStore.discardPendingDockLayout();
};

// Launchpad is reorderable but not removable: never let a drag-out gesture
// delete it, otherwise the icon becomes permanently unreachable.
const canRemoveOnDragOut = (
	item: DockerAppInfo | null
): item is DockerAppInfo =>
	!!item && item.id !== 'bdock:launchpad' && getAppIndexInDock(item.id) >= 0;

// Releasing a dock-origin drag outside the dock list unpins the app (drag-out
// to remove). Splice it out, then normalize the remaining icons to clean slots
// and persist the shared layout.
const removeOnDragOut = () => {
	const origin = origin_app;
	if (!canRemoveOnDragOut(origin)) {
		restoreReorder();
		return;
	}
	appStore.remove_app_on_docker(origin.id, false);
	reflowDockNormal();
	pushDockLayoutPC();
};

const onDockDrop = (e: any) => {
	e.preventDefault();
	dropHandled = true;

	// A reorder is finalized in onDragEnd (which always fires for a dock-origin
	// drag); here we only record that the release landed inside the dock.
	if (origin_app) {
		return;
	}

	// Launchpad -> dock. Prefer the prop, but fall back to the synchronous
	// dataTransfer payload so a fast drag (prop not yet emitted, dragKind still
	// 'none') still commits instead of dropping the icon on the floor.
	const raw = props.drag_launch_app || readDragText(e);
	if (!raw || !(raw.startsWith('bdesk:') || raw.startsWith('bdock:'))) {
		return;
	}
	if (pendingInsertIndex < 0) {
		pendingInsertIndex = computeInsertIndex(e).index;
	}
	// Mark the gesture so resetDragState() reflows the dock back to clean slots
	// even on a fast drag where dragover never promoted dragKind.
	dragKind = 'fromLaunchpad';
	commitLaunchpadDrop(raw);
	resetDragState();
};

const onDockDragLeave = (e: any) => {
	e.preventDefault();
	if (dragKind !== 'fromLaunchpad') return;
	// `dragleave` also fires when the pointer merely crosses from the list
	// container onto one of its child icons. Only collapse the preview when the
	// pointer truly leaves the dock list, otherwise the gap thrashes.
	const listEl = document.getElementById('app_dock_list');
	if (listEl && e.relatedTarget && listEl.contains(e.relatedTarget as Node)) {
		return;
	}
	reflowDockNormal();
	pendingInsertIndex = -1;
	dragPreviewActive.value = false;
};

// True only when the drag was released at a point geometrically outside the
// dock list. We can't infer drag-out from the mere absence of a `drop` event:
// browsers also skip `drop` on an Escape-key cancel or an accidental short
// drag that never leaves the dock, and treating those as drag-out would unpin
// an app the user never dragged off the dock.
const isReleaseOutsideDock = (e: any): boolean => {
	const listEl = document.getElementById('app_dock_list');
	if (!listEl) return false;
	const x = e?.clientX ?? 0;
	const y = e?.clientY ?? 0;
	// A cancelled drag (Escape) often reports 0,0 for the end coordinates;
	// never treat that as an off-dock release.
	if (x === 0 && y === 0) return false;
	const rect = listEl.getBoundingClientRect();
	return x < rect.left || x > rect.right || y < rect.top || y > rect.bottom;
};

const onDragEnd = (e: any) => {
	if (dragKind === 'reorder') {
		if (isReleaseOutsideDock(e)) {
			// Released off the dock => unpin (with the launchpad safeguard).
			removeOnDragOut();
		} else if (dropHandled) {
			// A real drop landed inside the dock list => commit the reorder.
			persistReorder();
		} else {
			// Released inside the dock without a drop (Escape-cancel or an
			// accidental drag): keep the app pinned, just restore its order.
			restoreReorder();
		}
	}
	resetDragState();
};

// When a launchpad drag ends without dropping on the dock, the dock receives
// no drop/dragend event; the cleared prop is our signal to collapse the
// preview gap and reset.
watch(
	() => props.drag_launch_app,
	(val) => {
		if (!val && dragKind === 'fromLaunchpad') {
			resetDragState();
		}
	}
);

const handleRemoveApp = (appid: string) => {
	const rid = normalizeDockId(appid);
	const cc = getAppIndexInDock(rid);
	if (cc < 0) return;
	if (appStore.dockerApps[cc].is_temp) {
		appStore.remove_app_on_docker(rid, true);
		return;
	}
	appStore.dockerApps[cc].show_dot = false;
	appStore.updateAppInDock(appStore.dockerApps[cc]);
};

const ensureDotVisible = (appid: string) => {
	const rid = normalizeDockId(appid);
	const cc = getAppIndexInDock(rid);
	if (cc < 0) return;
	appStore.dockerApps[cc].show_dot = true;
};

const updateAppTemp = (value: DockerAppInfo) => {
	const app = JSON.parse(JSON.stringify(value));
	app.is_temp = !app.is_temp;
	appStore.updateAppInDock(app);
	pushDockLayoutPC();
};

const removeAppInDock = (id: string) => {
	appStore.remove_app_on_docker(id, true);
	pushDockLayoutPC();
};

const handleOpenApp = (appid: string) => {
	const rid = normalizeDockId(appid);
	const cc = getAppIndexInDock(rid);
	if (cc < 0) {
		appStore.add_app_on_docker_bottom(rid);
	} else {
		appStore.dockerApps[cc].show_dot = true;
	}
};

const toSearch = () => {
	emits('changeSearchDialog', true);
};

const showNotification = () => {
	notificationStore.toggleNotificaitonDisplay();
};

defineExpose({
	handleRemoveApp,
	handleOpenApp,
	ensureDotVisible
});

const handleDocumentClick = () => {
	contextmenuId.value = '';
};

onMounted(() => {
	document.addEventListener('click', handleDocumentClick);
	void nextTick(() => {
		requestAnimationFrame(() => {
			transitionsEnabled.value = true;
		});
	});
});

onUnmounted(() => {
	document.removeEventListener('click', handleDocumentClick);
	appStore.setLayoutDragLock(false);
});
</script>

<style lang="scss" scoped>
.drag-enter {
	border: 1px dashed white;
}
.drag-start {
	opacity: 0;
}

.desktop-box {
	.desktop_app_box {
		width: 60px;
		height: 100%;
		display: flex;
		align-items: center;
		position: absolute;
		top: 0px;
		left: 24px;
		z-index: 999;
		.desktop_app {
			width: 100%;
			max-height: 94%;
			border: 1px solid rgba(255, 255, 255, 0.2);
			border-radius: 16px;
			padding-top: 20px;
			padding-bottom: 20px;

			background: rgba(246, 246, 246, 0.3);
			border: 1px solid rgba(255, 255, 255, 0.2);
			filter: drop-shadow(0px 0px 40px rgba(0, 0, 0, 0.2))
				drop-shadow(0px 0px 2px rgba(0, 0, 0, 0.4));
			backdrop-filter: blur(120px);
			display: flex;
			flex-direction: column;

			.desktop_avatar_border {
				width: 36px;
				height: 0px;
				opacity: 0.3;
				border-bottom: 1px solid rgba(31, 24, 20, 1);
			}
			.desktop_user_avatar {
				display: flex;
				flex-wrap: wrap;
				justify-content: center;
				.desktop_avatar {
					width: 36px;
					height: 36px;
					box-shadow: 0px 0px 11px 0px rgba(0, 0, 0, 0.2);
					border-radius: 50%;
					margin-bottom: 12px;
					cursor: pointer;
					overflow: hidden;
				}
				.dock_icon {
					margin-top: 14px;
					width: 32px;
					height: 32px;
					border-radius: 16px;
					// overflow: hidden;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					border-radius: 18px;
					background: rgba(255, 255, 255, 0.2);
					display: flex;
					align-items: center;
					justify-content: center;
					justify-content: center;
					position: relative;
					cursor: pointer;
					img {
						width: 16px;
						height: 16px;
					}

					.badge {
						width: 8px;
						height: 8px;
						border-radius: 4px;
						position: absolute;
						right: 0px;
						top: 0px;
						background-color: #fa473b;
					}
				}
				.desktop_search {
					margin-top: 14px;
					width: 100%;
					width: 32px;
					height: 32px;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					img {
						cursor: pointer;
					}
				}
				.desktop_notify {
					margin-top: 14px;
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					img {
						cursor: pointer;
					}
				}
				.desktop_logout {
					margin-top: 14px;
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					.desktop_message_icons {
						right: -17px;
					}
					img {
						cursor: pointer;
					}
				}
			}
			.column_apps_all {
				width: 100%;
				flex: 1;
				// min-height: 200px;
				// max-height: 555px;
				// overflow: hidden;
				position: relative;
				img {
					cursor: pointer;
				}
				.img_parent {
					display: flex;
					&::after {
						content: '';
						display: block;
						width: 4px;
						height: 4px;
						background: rgba(255, 255, 255, 0.8);
						border-radius: 50%;
						margin-right: 6px;
					}
				}
				.img_parent_minimized {
					display: flex;
					&::after {
						content: '';
						display: block;
						width: 4px;
						height: 4px;
						background: transparent;
						border: 1px solid rgba(255, 255, 255, 0.8);
						border-radius: 50%;
						margin-right: 6px;
						box-sizing: border-box;
					}
				}
				.menu_open {
					background: $background-1;
					// backdrop-filter: blur(120px);
					border-radius: 6px;
					padding: 4px;
					opacity: 0;
					animation: fadeIn 0.5s forwards;
					.context_triangle {
						position: absolute;
						top: 10px;
						left: -8px;
						width: 0;
						height: 0;
						border-top: 8px solid transparent;
						border-bottom: 8px solid transparent;
						border-right: 8px solid $background-1;
					}
					.item {
						font-size: 12px;
						line-height: 16px;
						font-family: Roboto-Regular, Roboto;
						font-weight: 400;
						color: $ink-1;
						text-align: left;
						margin: 0px;
						cursor: pointer;
						white-space: nowrap;
						padding: 4px 8px;
						border-radius: 4px;
						&:hover {
							background-color: rgba(0, 0, 0, 0.1);
						}

						&.red_font {
							color: $negative;
						}
					}
				}

				@keyframes fadeIn {
					to {
						opacity: 1;
					}
				}
				.menu_hide {
					display: none;
					opacity: 1;
					animation: fadeOut 0.5s forwards;
				}

				@keyframes fadeOut {
					to {
						opacity: 0;
					}
				}

				.column_myapps {
					width: 100%;
					display: flex;
					flex-wrap: wrap;
					justify-content: center;
					padding-top: 15px;
					cursor: pointer;
					.column_myapps_logo {
						width: 40px;
						height: 40px;
						overflow: hidden;
						box-shadow: 0px 2px 10px 0px rgba(0, 0, 0, 0.2);
						border-radius: 5px;
						margin-bottom: 16px;
						background-position: center;
						background-repeat: no-repeat;
						background-size: contain;
					}
				}
			}
		}
	}
}
</style>
