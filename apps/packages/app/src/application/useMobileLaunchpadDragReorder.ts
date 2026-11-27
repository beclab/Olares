import { ref, watch, nextTick, type Ref } from 'vue';
import { localStoreKeys, useApplicationStore } from 'src/stores/desktop/app';
import {
	LAUNCHPAD_GRID_ROW_GAP_PX,
	LAUNCHPAD_GRID_ROW_TRACK_PX
} from 'src/application/mobile';
import {
	LAYOUT_SCHEMA_VERSION,
	LaunchLayout,
	LaunchpadItem,
	rawIdToKey
} from 'src/types/desktop-layout';

const LONG_PRESS_MS = 400;
const MOVE_THRESHOLD_PX = 6;
const SWIPE_COMMIT_PX = 8;
const SWIPE_DIRECTION_BIAS = 1.5;
const PICKED_SCALE = 1.08;
const EDGE_FLIP_DWELL_MS = 520;
const EDGE_FLIP_COOLDOWN_MS = 720;
const EDGE_LEAVE_GRACE_MS = 200;
const MOBILE_GRID_HIT_MARGIN = 6;
const MOBILE_EDGE_TOUCH_TOLERANCE = 2;
const POINTER_CAPTURE_CLICK_GUARD_MS = 300;
const VIBRATE_MS = 15;
const EDGE_TRIGGER_PX = 36;
const FALLBACK_PARENT_WALK_DEPTH = 5;
const LS_IDLE_TIMEOUT_MS = 1500;

type Stage = 'idle' | 'pressing' | 'picked' | 'dragging';

type DragSnapshotItem = {
	id: string;
	index: number;
	page: number;
	page_num: number;
	left: number;
	top: number;
};

function copyComputedSize(srcEl: HTMLElement, dstEl: HTMLElement) {
	const cs = getComputedStyle(srcEl);
	dstEl.style.width = cs.width;
	dstEl.style.height = cs.height;
	dstEl.style.fontSize = cs.fontSize;
	dstEl.style.lineHeight = cs.lineHeight;
	dstEl.style.color = cs.color;
	dstEl.style.fontFamily = cs.fontFamily;
	dstEl.style.fontWeight = cs.fontWeight;
	dstEl.style.textAlign = cs.textAlign;
	dstEl.style.borderRadius = cs.borderRadius;
}

function createMobileGhost(
	srcEl: HTMLElement,
	mountRoot: HTMLElement | null
): { host: HTMLElement } {
	const rect = srcEl.getBoundingClientRect();

	const clone = srcEl.cloneNode(true) as HTMLElement;
	clone.removeAttribute('id');
	clone.querySelectorAll('.delete_launch').forEach((el) => el.remove());
	clone
		.querySelectorAll('.install_loading_status')
		.forEach((el) => el.remove());
	clone.style.position = 'relative';
	clone.style.top = '0';
	clone.style.left = '0';
	clone.style.width = `${rect.width}px`;
	clone.style.height = `${rect.height}px`;
	clone.style.margin = '0';
	clone.style.animation = 'none';
	clone.style.webkitAnimation = 'none';
	clone.style.transform = 'none';
	clone.style.transition = 'none';
	clone.style.opacity = '1';

	const srcImg = srcEl.querySelector('img.pad-img') as HTMLElement | null;
	const dstImg = clone.querySelector('img.pad-img') as HTMLElement | null;
	if (srcImg && dstImg) copyComputedSize(srcImg, dstImg);

	const srcLabel = srcEl.querySelector(
		'.launchpadapps_name'
	) as HTMLElement | null;
	const dstLabel = clone.querySelector(
		'.launchpadapps_name'
	) as HTMLElement | null;
	if (srcLabel && dstLabel) copyComputedSize(srcLabel, dstLabel);

	const shell = document.createElement('div');
	shell.className = 'lp-mobile-drag-ghost';
	shell.style.position = 'fixed';
	shell.style.left = '0';
	shell.style.top = '0';
	shell.style.width = `${rect.width}px`;
	shell.style.height = `${rect.height}px`;
	shell.style.pointerEvents = 'none';
	shell.style.touchAction = 'none';
	shell.style.zIndex = '2147483647';
	shell.style.willChange = 'transform, opacity';
	shell.style.transition = 'transform 120ms ease-out';
	shell.style.opacity = '0';
	shell.style.transformOrigin = 'center center'; // i18n-ignore
	shell.style.boxSizing = 'border-box';
	shell.appendChild(clone);

	const mount = mountRoot ?? document.body;
	mount.appendChild(shell);

	return { host: shell };
}

export function useMobileLaunchpadDragReorder(opts: {
	slide: Ref<number>;
	carousel: Ref<{ goTo?: (n: number) => void } | undefined>;
	carouselAnimated?: Ref<boolean>;
	launchpadMount?: Ref<HTMLElement | null | undefined>;
	appStore: ReturnType<typeof useApplicationStore>;
	isDisplay: Ref<boolean>;
	exitSearchForDrag?: (draggedAppId?: string) => void;
	applyPendingSlideRestoreAfterDrag?: () => number | null;
	clearPendingSlideAfterSearchDrag?: () => void;
	resumeCarouselAfterSearchDrag?: () => void;
}) {
	const {
		slide,
		carousel,
		carouselAnimated,
		launchpadMount,
		appStore,
		isDisplay,
		exitSearchForDrag,
		applyPendingSlideRestoreAfterDrag,
		clearPendingSlideAfterSearchDrag,
		resumeCarouselAfterSearchDrag
	} = opts;

	const dragLaunchAppId = ref<string | null>(null);
	const isInDrag = ref(false);
	const lastDragFinishTime = ref(0);

	let stage: Stage = 'idle';
	let pressTimer: ReturnType<typeof setTimeout> | null = null;
	let activePointerId: number | null = null;
	let sourceEl: HTMLElement | null = null;
	let sourceAppId: string | null = null;
	let pressStartX = 0;
	let pressStartY = 0;
	let lastPointerX = 0;
	let lastPointerY = 0;
	let pointerOffsetX = 0;
	let pointerOffsetY = 0;
	let ghostHost: HTMLElement | null = null;
	let choose = -1;
	let lastReorderTargetIndex: number | null = null;

	let preDragSnapshot: DragSnapshotItem[] | null = null;
	let pendingSearchExitOnDragInit = false;
	let dragStartGeneration = 0;
	let launchPadSearchExitRaf = 0;

	let edgeFlipDwellTimer: ReturnType<typeof setTimeout> | null = null;
	let edgeFlipPendingDir: -1 | 0 | 1 = 0;
	let lastEdgeFlipAt = 0;
	let edgeRepeatTimer: ReturnType<typeof setTimeout> | null = null;
	let edgeLeaveGraceTimer: ReturnType<typeof setTimeout> | null = null;

	let dragRafId = 0;
	let lastTickX = NaN;
	let lastTickY = NaN;

	let cachedSlideForCache = -1;
	let cachedSlideSurfaceEl: HTMLElement | null = null;
	let cachedSlideRect: DOMRect | null = null;
	let cachedViewportBounds: {
		left: number;
		right: number;
		width: number;
	} | null = null;
	let resizeListenerAttached = false;

	let pendingLsPayload: string | null = null;
	let pendingLsIdleId = 0;
	let pendingLsIsTimer = false;

	function getCarouselRootEl(): HTMLElement | null {
		const c = carousel.value as { $el?: HTMLElement } | HTMLElement | undefined;
		if (!c) return null;
		if (c instanceof HTMLElement) return c;
		const el = c.$el;
		return el instanceof HTMLElement ? el : null;
	}

	function computeEdgeFlipScreenBounds(): {
		left: number;
		right: number;
		width: number;
	} {
		const iw = window.innerWidth;
		const vv = window.visualViewport;
		if (vv && vv.width > 0) {
			const vvRight = vv.offsetLeft + vv.width;
			const left = Math.min(0, vv.offsetLeft);
			const right = Math.max(iw, vvRight);
			return { left, right, width: Math.max(1, right - left) };
		}
		return { left: 0, right: iw, width: Math.max(1, iw) };
	}

	function getCachedViewportBounds(): {
		left: number;
		right: number;
		width: number;
	} {
		if (!cachedViewportBounds) {
			cachedViewportBounds = computeEdgeFlipScreenBounds();
		}
		return cachedViewportBounds;
	}

	function computeSlideSurfaceEl(slideIndex: number): HTMLElement | null {
		const root = getCarouselRootEl();
		if (!root) return null;
		return root.querySelector(
			`[data-lp-slide="${slideIndex}"]`
		) as HTMLElement | null;
	}

	function getCachedSlideSurfaceEl(): HTMLElement | null {
		if (cachedSlideForCache !== slide.value || !cachedSlideSurfaceEl) {
			cachedSlideSurfaceEl = computeSlideSurfaceEl(slide.value);
			cachedSlideForCache = slide.value;
			cachedSlideRect = null;
		}
		return cachedSlideSurfaceEl;
	}

	function getCachedSlideRect(): DOMRect | null {
		const el = getCachedSlideSurfaceEl();
		if (!el) return null;
		return el.getBoundingClientRect();
	}

	function invalidateSurfaceCache() {
		cachedSlideForCache = -1;
		cachedSlideSurfaceEl = null;
		cachedSlideRect = null;
	}

	function invalidateViewportCache() {
		cachedViewportBounds = null;
	}

	function onWindowResize() {
		invalidateSurfaceCache();
		invalidateViewportCache();
	}

	function attachResizeListener() {
		if (resizeListenerAttached) return;
		window.addEventListener('resize', onWindowResize);
		window.visualViewport?.addEventListener('resize', onWindowResize);
		resizeListenerAttached = true;
	}

	function detachResizeListener() {
		if (!resizeListenerAttached) return;
		window.removeEventListener('resize', onWindowResize);
		window.visualViewport?.removeEventListener('resize', onWindowResize);
		resizeListenerAttached = false;
	}

	function isLaunchpadSlideMounted(slideIndex: number): boolean {
		if (slide.value !== slideIndex) return false;
		const surface = computeSlideSurfaceEl(slideIndex);
		return surface != null && surface.getBoundingClientRect().width > 0;
	}

	function syncCarouselToSlide(target: number, dragStartGen: number) {
		if (carouselAnimated) {
			carouselAnimated.value = false;
		}

		let attempts = 0;
		let usedResetBump = false;

		const trySync = () => {
			if (dragStartGen !== dragStartGeneration) return;
			attempts++;

			const pagesReady = appStore.launchPadApps.length > target;
			if (!pagesReady && attempts < 32) {
				void nextTick(() => trySync());
				return;
			}

			if (
				!usedResetBump &&
				target !== 0 &&
				attempts > 2 &&
				!isLaunchpadSlideMounted(target)
			) {
				usedResetBump = true;
				slide.value = 0;
				void nextTick(() => {
					if (dragStartGen !== dragStartGeneration) return;
					slide.value = target;
					carousel.value?.goTo?.(target);
					void nextTick(() => trySync());
				});
				return;
			}

			slide.value = target;
			carousel.value?.goTo?.(target);

			const mounted = isLaunchpadSlideMounted(target);
			if (!mounted && attempts < 32) {
				void nextTick(() => trySync());
			}
		};

		void nextTick(() => {
			void nextTick(() => trySync());
		});
	}

	function snapshotDesktopApps(): DragSnapshotItem[] {
		return appStore.desktopApps.map((a) => ({
			id: a.id,
			index: a.index,
			page: a.page,
			page_num: a.page_num,
			left: a.left,
			top: a.top
		}));
	}

	function restoreFromSnapshot() {
		if (!preDragSnapshot) return;
		const byId = new Map(preDragSnapshot.map((s) => [s.id, s]));
		for (const app of appStore.desktopApps) {
			const snap = byId.get(app.id);
			if (!snap) continue;
			app.index = snap.index;
			app.page = snap.page;
			app.page_num = snap.page_num;
			app.left = snap.left;
			app.top = snap.top;
		}
		rebuildLaunchPadApps();
	}

	function applySearchExitAfterDragReady(
		startId: string,
		dragStartGen: number
	) {
		launchPadSearchExitRaf = requestAnimationFrame(() => {
			launchPadSearchExitRaf = 0;
			if (dragStartGen !== dragStartGeneration) return;
			if (!dragLaunchAppId.value) return;

			exitSearchForDrag?.(startId);
			pendingSearchExitOnDragInit = false;

			preDragSnapshot = snapshotDesktopApps();

			void nextTick(() => {
				if (dragStartGen !== dragStartGeneration) return;
				if (!dragLaunchAppId.value) return;

				const restoredSlide = applyPendingSlideRestoreAfterDrag?.();
				if (restoredSlide != null) {
					syncCarouselToSlide(restoredSlide, dragStartGen);
				}
				choose = getAppIndexInDesktop(startId);
			});
		});
	}

	function getPerPage() {
		const gridX =
			appStore.launchpadGridOverride?.x ?? appStore.DESKTOP_APP_X_NUM;
		const gridY =
			appStore.launchpadGridOverride?.y ?? appStore.DESKTOP_APP_Y_NUM;
		return Math.max(1, gridX * gridY);
	}

	function rebuildLaunchPadApps() {
		if (appStore.desktopApps.length === 0) {
			appStore.launchPadApps = [[]];
			return;
		}
		let maxPage = 0;
		for (const a of appStore.desktopApps) {
			if (a.page > maxPage) maxPage = a.page;
		}
		const pages: number[][] = Array.from({ length: maxPage + 1 }, () => []);
		for (let i = 0; i < appStore.desktopApps.length; i++) {
			pages[appStore.desktopApps[i].page].push(i);
		}
		for (const page of pages) {
			page.sort(
				(a, b) =>
					appStore.desktopApps[a].page_num - appStore.desktopApps[b].page_num
			);
		}
		appStore.launchPadApps = pages;
		while (
			appStore.launchPadApps.length > 1 &&
			appStore.launchPadApps[appStore.launchPadApps.length - 1].length === 0
		) {
			appStore.launchPadApps.pop();
		}
	}

	function applyReorderInPlace(toIndex: number) {
		if (choose < 0 || choose >= appStore.desktopApps.length) return;
		const dragged = appStore.desktopApps[choose];
		const fromIndex = dragged.index;
		if (fromIndex === toIndex) return;

		const perPage = getPerPage();
		const oldPage = dragged.page;
		const newPage = Math.floor(toIndex / perPage);

		appStore.$patch(() => {
			if (fromIndex < toIndex) {
				for (const a of appStore.desktopApps) {
					if (a === dragged) continue;
					if (a.index > fromIndex && a.index <= toIndex) {
						a.index -= 1;
						a.page = Math.floor(a.index / perPage);
						a.page_num = a.index % perPage;
					}
				}
			} else {
				for (const a of appStore.desktopApps) {
					if (a === dragged) continue;
					if (a.index < fromIndex && a.index >= toIndex) {
						a.index += 1;
						a.page = Math.floor(a.index / perPage);
						a.page_num = a.index % perPage;
					}
				}
			}
			dragged.index = toIndex;
			dragged.page = newPage;
			dragged.page_num = toIndex % perPage;

			const pages = appStore.launchPadApps;
			while (newPage >= pages.length) {
				pages.push([]);
			}

			const lo = Math.min(oldPage, newPage);
			const hi = Math.max(oldPage, newPage);
			for (let p = lo; p <= hi; p++) {
				if (p >= pages.length) continue;
				const newPageArr: number[] = [];
				for (let i = 0; i < appStore.desktopApps.length; i++) {
					if (appStore.desktopApps[i].page === p) {
						newPageArr.push(i);
					}
				}
				newPageArr.sort(
					(a, b) =>
						appStore.desktopApps[a].page_num - appStore.desktopApps[b].page_num
				);
				pages[p].splice(0, pages[p].length, ...newPageArr);
			}
		});
	}

	function getAppIndexInDesktop(id: string) {
		return appStore.desktopApps.findIndex((a) => a.id === id);
	}

	function getGhostBounds(): DOMRect | null {
		return ghostHost?.getBoundingClientRect() ?? null;
	}

	function getReorderPointerCoords(): { x: number; y: number } {
		const ghost = getGhostBounds();
		if (ghost) {
			return {
				x: ghost.left + ghost.width / 2,
				y: ghost.top + ghost.height / 2
			};
		}
		return { x: lastPointerX, y: lastPointerY };
	}

	function pointerToTargetMobileIndex(
		clientX: number,
		clientY: number
	): number | null {
		const rect = getCachedSlideRect();
		if (!rect) {
			return null;
		}

		const margin = MOBILE_GRID_HIT_MARGIN;
		if (
			clientX < rect.left - margin ||
			clientX > rect.right + margin ||
			clientY < rect.top - margin ||
			clientY > rect.bottom + margin
		) {
			return null;
		}

		const lx = clientX - rect.left;
		const ly = clientY - rect.top;

		const gridX =
			appStore.launchpadGridOverride?.x ?? appStore.DESKTOP_APP_X_NUM;
		const gridY =
			appStore.launchpadGridOverride?.y ?? appStore.DESKTOP_APP_Y_NUM;
		const perPage = gridX * gridY;
		const n = appStore.desktopApps.length;
		if (n === 0) return null;

		const colW = rect.width / gridX;
		const rowStride = LAUNCHPAD_GRID_ROW_TRACK_PX + LAUNCHPAD_GRID_ROW_GAP_PX;

		let col = Math.floor(lx / colW);
		let row = Math.floor(ly / rowStride);
		col = Math.max(0, Math.min(gridX - 1, col));
		row = Math.max(0, Math.min(gridY - 1, row));

		const slot = row * gridX + col;
		const raw = slide.value * perPage + slot;
		return Math.max(0, Math.min(n - 1, raw));
	}

	function maybeApplyReorderFromPointer(clientX: number, clientY: number) {
		if (!dragLaunchAppId.value || choose < 0) return;
		if (
			appStore.launchpadSearchQuery.length > 0 ||
			pendingSearchExitOnDragInit
		) {
			return;
		}

		const { x, y } = getReorderPointerCoords();
		const probeX = ghostHost ? x : clientX;
		const probeY = ghostHost ? y : clientY;

		const fromGrid = pointerToTargetMobileIndex(probeX, probeY);
		if (fromGrid !== null) {
			if (lastReorderTargetIndex === fromGrid) return;
			lastReorderTargetIndex = fromGrid;
			applyReorderInPlace(fromGrid);
			return;
		}

		const el = document.elementFromPoint(probeX, probeY);
		let cur: HTMLElement | null = el as HTMLElement | null;
		let depth = 0;
		while (
			cur &&
			depth < FALLBACK_PARENT_WALK_DEPTH &&
			(!cur.id || !cur.id.startsWith('bdesk:'))
		) {
			cur = cur.parentElement;
			depth++;
		}
		if (
			cur &&
			cur.id &&
			cur.id.startsWith('bdesk:') &&
			cur.id !== dragLaunchAppId.value
		) {
			const hitIdx = getAppIndexInDesktop(cur.id);
			if (hitIdx >= 0) {
				const toIdx = appStore.desktopApps[hitIdx].index;
				if (lastReorderTargetIndex === toIdx) return;
				lastReorderTargetIndex = toIdx;
				applyReorderInPlace(toIdx);
				return;
			}
		}

		lastReorderTargetIndex = null;
	}

	function cancelEdgeLeaveGrace() {
		if (edgeLeaveGraceTimer != null) {
			clearTimeout(edgeLeaveGraceTimer);
			edgeLeaveGraceTimer = null;
		}
	}

	function scheduleEdgeLeaveGraceClear() {
		if (edgeLeaveGraceTimer != null) return;
		edgeLeaveGraceTimer = setTimeout(() => {
			edgeLeaveGraceTimer = null;
			clearEdgeFlipDwell();
		}, EDGE_LEAVE_GRACE_MS);
	}

	function clearEdgeRepeatProbe() {
		if (edgeRepeatTimer != null) {
			clearTimeout(edgeRepeatTimer);
			edgeRepeatTimer = null;
		}
	}

	function scheduleEdgeRepeatProbe() {
		clearEdgeRepeatProbe();
		edgeRepeatTimer = setTimeout(() => {
			edgeRepeatTimer = null;
			if (!dragLaunchAppId.value) return;
			void nextTick(() => {
				if (!dragLaunchAppId.value) return;
				maybeFlipPageFromGhost();
			});
		}, EDGE_FLIP_COOLDOWN_MS);
	}

	function clearEdgeFlipDwell() {
		cancelEdgeLeaveGrace();
		if (edgeFlipDwellTimer != null) {
			clearTimeout(edgeFlipDwellTimer);
			edgeFlipDwellTimer = null;
		}
		edgeFlipPendingDir = 0;
	}

	function getGhostEdgeFlipDir(): -1 | 0 | 1 {
		const ghost = getGhostBounds();
		if (!ghost) return 0;

		const screen = getCachedViewportBounds();
		const pad = MOBILE_EDGE_TOUCH_TOLERANCE;

		if (ghost.left <= screen.left + pad && slide.value > 0) return -1;
		if (
			ghost.right >= screen.right - pad &&
			slide.value < appStore.launchPadApps.length - 1
		) {
			return 1;
		}
		return 0;
	}

	function maybeFlipPageFromGhost() {
		if (!dragLaunchAppId.value) {
			clearEdgeFlipDwell();
			clearEdgeRepeatProbe();
			return;
		}
		if (appStore.launchPadApps.length <= 1) {
			clearEdgeFlipDwell();
			clearEdgeRepeatProbe();
			return;
		}

		const vp = getCachedViewportBounds();
		const nearLeft = lastPointerX < vp.left + EDGE_TRIGGER_PX;
		const nearRight = lastPointerX > vp.right - EDGE_TRIGGER_PX;
		if (!nearLeft && !nearRight) {
			if (!edgeFlipDwellTimer) {
				scheduleEdgeLeaveGraceClear();
			}
			return;
		}

		if (Date.now() - lastEdgeFlipAt < EDGE_FLIP_COOLDOWN_MS) {
			return;
		}

		const dir = getGhostEdgeFlipDir();

		if (dir === 0) {
			if (!edgeFlipDwellTimer) {
				scheduleEdgeLeaveGraceClear();
			}
			return;
		}

		cancelEdgeLeaveGrace();

		if (edgeFlipPendingDir !== dir) {
			clearEdgeRepeatProbe();
			clearEdgeFlipDwell();
			edgeFlipPendingDir = dir;
			const scheduledDir = dir;
			edgeFlipDwellTimer = setTimeout(() => {
				edgeFlipDwellTimer = null;
				edgeFlipPendingDir = 0;
				if (!dragLaunchAppId.value) return;
				if (Date.now() - lastEdgeFlipAt < EDGE_FLIP_COOLDOWN_MS) return;

				const ok = getGhostEdgeFlipDir();
				if (ok === scheduledDir) {
					if (carouselAnimated) carouselAnimated.value = true;
					carousel.value?.goTo?.(slide.value + scheduledDir);
					lastEdgeFlipAt = Date.now();
					scheduleEdgeRepeatProbe();
				}
			}, EDGE_FLIP_DWELL_MS);
		}
	}

	function applyGhostTransform(clientX: number, clientY: number) {
		if (!ghostHost) return;
		const x = clientX - pointerOffsetX;
		const y = clientY - pointerOffsetY;
		ghostHost.style.transform = `translate3d(${x}px, ${y}px, 0) scale(${PICKED_SCALE})`;
	}

	function spawnGhost(clientX: number, clientY: number) {
		if (!sourceEl) return;
		try {
			const r = createMobileGhost(sourceEl, launchpadMount?.value ?? null);
			ghostHost = r.host;
			applyGhostTransform(clientX, clientY);
			requestAnimationFrame(() => {
				if (!ghostHost) return;
				ghostHost.style.opacity = '1';
			});
		} catch (err) {
			console.warn('[mobile-launchpad-drag] ghost spawn failed', err);
		}
	}

	function disposeGhost() {
		if (ghostHost?.isConnected) {
			ghostHost.remove();
		}
		ghostHost = null;
	}

	function dragTick() {
		dragRafId = 0;
		if (stage !== 'picked' && stage !== 'dragging') return;

		if (stage === 'picked') {
			const dx = lastPointerX - pressStartX;
			const dy = lastPointerY - pressStartY;
			if (Math.hypot(dx, dy) > MOVE_THRESHOLD_PX) {
				stage = 'dragging';
			}
		}

		if (lastTickX !== lastPointerX || lastTickY !== lastPointerY) {
			applyGhostTransform(lastPointerX, lastPointerY);
			if (stage === 'dragging') {
				maybeApplyReorderFromPointer(lastPointerX, lastPointerY);
				maybeFlipPageFromGhost();
			}
			lastTickX = lastPointerX;
			lastTickY = lastPointerY;
		}

		dragRafId = requestAnimationFrame(dragTick);
	}

	function startDragTickLoop() {
		if (dragRafId !== 0) return;
		lastTickX = NaN;
		lastTickY = NaN;
		dragRafId = requestAnimationFrame(dragTick);
	}

	function stopDragTickLoop() {
		if (dragRafId !== 0) {
			cancelAnimationFrame(dragRafId);
			dragRafId = 0;
		}
		lastTickX = NaN;
		lastTickY = NaN;
	}

	function pushLaunchLayoutMobile() {
		const cur = appStore.layoutBySurface.launchpad;
		const sorted = [...appStore.desktopApps].sort((a, b) => a.index - b.index);
		const items: LaunchpadItem[] = [];
		const seen = new Set<string>();
		for (let i = 0; i < sorted.length; i++) {
			const raw = (sorted[i] as { id: string }).id;
			const rawId = raw.startsWith('bdesk:') ? raw.slice(6) : raw;
			if (!rawId || seen.has(rawId)) continue;
			seen.add(rawId);
			items.push({ key: rawIdToKey(rawId), order: items.length });
		}
		const next: LaunchLayout = {
			schemaVersion: LAYOUT_SCHEMA_VERSION,
			items,
			hidden: cur?.hidden ?? [],
			version: cur?.version ?? 0,
			updatedAt: Date.now(),
			updatedBy: ''
		};
		appStore.saveLaunchLayoutDebounced(next);
	}

	function schedulePendingLs(payload: string) {
		cancelPendingLs();
		pendingLsPayload = payload;

		const doWrite = () => {
			pendingLsIdleId = 0;
			if (pendingLsPayload != null) {
				try {
					localStorage.setItem(localStoreKeys.LAUNCH_PAD, pendingLsPayload);
				} catch (err) {
					console.warn(
						'[mobile-launchpad-drag] localStorage write failed',
						err
					);
				}
				pendingLsPayload = null;
				pushLaunchLayoutMobile();
			}
		};

		const ric = (
			window as unknown as {
				requestIdleCallback?: (
					cb: () => void,
					opts?: { timeout?: number }
				) => number;
			}
		).requestIdleCallback;
		if (typeof ric === 'function') {
			pendingLsIdleId = ric(doWrite, { timeout: LS_IDLE_TIMEOUT_MS });
			pendingLsIsTimer = false;
		} else {
			pendingLsIdleId = setTimeout(doWrite, 0) as unknown as number;
			pendingLsIsTimer = true;
		}
	}

	function cancelPendingLs() {
		if (pendingLsIdleId === 0) return;
		const cic = (
			window as unknown as {
				cancelIdleCallback?: (id: number) => void;
			}
		).cancelIdleCallback;
		if (pendingLsIsTimer) {
			clearTimeout(pendingLsIdleId);
		} else if (typeof cic === 'function') {
			cic(pendingLsIdleId);
		}
		pendingLsIdleId = 0;
	}

	function flushPendingLs() {
		if (pendingLsPayload == null) {
			cancelPendingLs();
			return;
		}
		cancelPendingLs();
		try {
			localStorage.setItem(localStoreKeys.LAUNCH_PAD, pendingLsPayload);
		} catch (err) {
			console.warn('[mobile-launchpad-drag] localStorage flush failed', err);
		}
		pendingLsPayload = null;
		pushLaunchLayoutMobile();
	}

	function blurFocusedInputForDrag() {
		if (typeof document === 'undefined') return;
		const ae = document.activeElement as HTMLElement | null;
		if (!ae || ae === document.body) return;
		const tag = ae.tagName;
		const isEditable =
			tag === 'INPUT' || tag === 'TEXTAREA' || ae.isContentEditable === true;
		if (!isEditable) return;
		try {
			ae.blur();
		} catch {
			/* noop */
		}
	}

	function transitionToPicked() {
		if (!sourceAppId) return;

		blurFocusedInputForDrag();

		const startId = sourceAppId;
		const dragStartGen = ++dragStartGeneration;
		const deferSearchExit = pendingSearchExitOnDragInit;
		const enteringEditMode = !isDisplay.value;

		stage = 'picked';
		isDisplay.value = true;
		isInDrag.value = true;
		dragLaunchAppId.value = startId;
		lastReorderTargetIndex = null;

		if (!deferSearchExit) {
			preDragSnapshot = snapshotDesktopApps();
			choose = getAppIndexInDesktop(startId);
		} else {
			choose = -1;
		}

		if (sourceEl && activePointerId != null) {
			try {
				sourceEl.setPointerCapture(activePointerId);
			} catch {
				/* noop */
			}
		}

		spawnGhost(lastPointerX, lastPointerY);

		invalidateSurfaceCache();
		invalidateViewportCache();
		attachResizeListener();
		attachTouchBlocker();
		startDragTickLoop();

		if (enteringEditMode) {
			try {
				if (typeof navigator !== 'undefined' && navigator.vibrate) {
					navigator.vibrate(VIBRATE_MS);
				}
			} catch {
				/* noop */
			}
		}

		if (deferSearchExit) {
			applySearchExitAfterDragReady(startId, dragStartGen);
		} else {
			const restoredSlide = applyPendingSlideRestoreAfterDrag?.();
			if (restoredSlide != null) {
				syncCarouselToSlide(restoredSlide, dragStartGen);
			}
		}
	}

	function onDocPointerMove(e: PointerEvent) {
		if (activePointerId !== e.pointerId) return;
		lastPointerX = e.clientX;
		lastPointerY = e.clientY;

		if (stage === 'pressing') {
			const dx = e.clientX - pressStartX;
			const dy = e.clientY - pressStartY;
			if (Math.hypot(dx, dy) > SWIPE_COMMIT_PX) {
				cancelPress();
				lastDragFinishTime.value = Date.now();
			}
			return;
		}

		if (stage === 'picked' || stage === 'dragging') {
			if (e.cancelable) e.preventDefault();
		}
	}

	function commitDrag() {
		rebuildLaunchPadApps();
		const apps = appStore.desktopApps
			.map((a) => ({ id: a.id, index: a.index }))
			.sort((x, y) => x.index - y.index);
		schedulePendingLs(JSON.stringify(apps));

		const maxSlide = Math.max(0, appStore.launchPadApps.length - 1);
		if (slide.value > maxSlide) {
			slide.value = maxSlide;
			void nextTick(() => {
				carousel.value?.goTo?.(maxSlide);
			});
		}
	}

	function finalizeRelease() {
		const wasPicked = stage === 'picked' || stage === 'dragging';
		const wasDragging = stage === 'dragging';

		stopDragTickLoop();

		resumeCarouselAfterSearchDrag?.();

		if (pendingSearchExitOnDragInit) {
			exitSearchForDrag?.(dragLaunchAppId.value ?? undefined);
			pendingSearchExitOnDragInit = false;
		}

		if (wasDragging) {
			commitDrag();
			lastDragFinishTime.value = Date.now();
		} else if (wasPicked && preDragSnapshot) {
			restoreFromSnapshot();
			lastDragFinishTime.value = Date.now();
		} else if (wasPicked) {
			lastDragFinishTime.value = Date.now();
		}

		if (sourceEl && activePointerId != null) {
			try {
				sourceEl.releasePointerCapture(activePointerId);
			} catch {
				/* noop */
			}
		}

		disposeGhost();
		clearEdgeFlipDwell();
		clearEdgeRepeatProbe();
		detachResizeListener();
		detachTouchBlocker();
		invalidateSurfaceCache();
		invalidateViewportCache();

		isInDrag.value = false;
		dragLaunchAppId.value = null;
		choose = -1;
		lastReorderTargetIndex = null;
		preDragSnapshot = null;
		sourceEl = null;
		sourceAppId = null;
		activePointerId = null;
		stage = 'idle';
	}

	function onDocPointerUp(e: PointerEvent) {
		if (activePointerId !== e.pointerId) return;

		if (pressTimer != null) {
			clearTimeout(pressTimer);
			pressTimer = null;
		}

		cleanupDocListeners();
		finalizeRelease();
	}

	function cancelPress() {
		if (pressTimer != null) {
			clearTimeout(pressTimer);
			pressTimer = null;
		}
		pendingSearchExitOnDragInit = false;
		clearPendingSlideAfterSearchDrag?.();
		cleanupDocListeners();
		sourceEl = null;
		sourceAppId = null;
		activePointerId = null;
		stage = 'idle';
	}

	function onDocTouchMoveBlock(e: TouchEvent) {
		if (stage !== 'picked' && stage !== 'dragging') return;
		if (e.cancelable) e.preventDefault();
		e.stopImmediatePropagation();
	}

	let touchBlockerAttached = false;
	function attachTouchBlocker() {
		if (touchBlockerAttached) return;
		document.addEventListener('touchmove', onDocTouchMoveBlock, {
			capture: true,
			passive: false
		});
		touchBlockerAttached = true;
	}
	function detachTouchBlocker() {
		if (!touchBlockerAttached) return;
		document.removeEventListener('touchmove', onDocTouchMoveBlock, true);
		touchBlockerAttached = false;
	}

	function attachDocListeners() {
		document.addEventListener('pointermove', onDocPointerMove, {
			capture: true,
			passive: false
		});
		document.addEventListener('pointerup', onDocPointerUp, { capture: true });
		document.addEventListener('pointercancel', onDocPointerUp, {
			capture: true
		});
	}

	function cleanupDocListeners() {
		document.removeEventListener('pointermove', onDocPointerMove, true);
		document.removeEventListener('pointerup', onDocPointerUp, true);
		document.removeEventListener('pointercancel', onDocPointerUp, true);
	}

	function onTilePointerDown(e: PointerEvent, appId: string) {
		if (stage !== 'idle') return;
		if (e.pointerType === 'mouse' && e.button !== 0) return;

		const target =
			(e.currentTarget as HTMLElement | null) ??
			(e.target as HTMLElement | null);
		if (!target) return;

		pendingSearchExitOnDragInit = appStore.launchpadSearchQuery.length > 0;
		preDragSnapshot = null;

		sourceEl = target;
		sourceAppId = appId;
		activePointerId = e.pointerId;
		pressStartX = e.clientX;
		pressStartY = e.clientY;
		lastPointerX = e.clientX;
		lastPointerY = e.clientY;

		const rect = target.getBoundingClientRect();
		pointerOffsetX = Math.max(0, Math.round(e.clientX - rect.left));
		pointerOffsetY = Math.max(0, Math.round(e.clientY - rect.top));

		stage = 'pressing';

		attachDocListeners();

		pressTimer = setTimeout(() => {
			pressTimer = null;
			if (stage !== 'pressing') return;
			transitionToPicked();
		}, LONG_PRESS_MS);
	}

	function shouldSuppressClick(): boolean {
		return (
			Date.now() - lastDragFinishTime.value < POINTER_CAPTURE_CLICK_GUARD_MS
		);
	}

	watch(slide, () => {
		invalidateSurfaceCache();
		if (!dragLaunchAppId.value) return;
		void nextTick(() => {
			requestAnimationFrame(() => {
				if (!dragLaunchAppId.value) return;
				maybeApplyReorderFromPointer(lastPointerX, lastPointerY);
			});
		});
	});

	function dispose() {
		if (pressTimer != null) {
			clearTimeout(pressTimer);
			pressTimer = null;
		}
		if (launchPadSearchExitRaf) {
			cancelAnimationFrame(launchPadSearchExitRaf);
			launchPadSearchExitRaf = 0;
		}
		dragStartGeneration++;
		clearPendingSlideAfterSearchDrag?.();
		pendingSearchExitOnDragInit = false;
		cleanupDocListeners();
		detachTouchBlocker();
		clearEdgeFlipDwell();
		clearEdgeRepeatProbe();
		stopDragTickLoop();
		detachResizeListener();
		flushPendingLs();
		invalidateSurfaceCache();
		invalidateViewportCache();
		disposeGhost();
		isInDrag.value = false;
		dragLaunchAppId.value = null;
		choose = -1;
		preDragSnapshot = null;
		sourceEl = null;
		sourceAppId = null;
		activePointerId = null;
		stage = 'idle';
	}

	return {
		dragLaunchAppId,
		isInDrag,
		lastDragFinishTime,
		onTilePointerDown,
		shouldSuppressClick,
		dispose
	};
}
