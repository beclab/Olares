import { ref, watch, nextTick, type Ref } from 'vue';
import { localStoreKeys, useApplicationStore } from 'src/stores/desktop/app';
import type { DesktopPosition } from 'src/apps/Desktop/type/types';
import {
	LAYOUT_SCHEMA_VERSION,
	LaunchLayout,
	LaunchpadItem,
	rawIdToKey
} from 'src/types/desktop-layout';

type DragSnapshotItem = {
	id: string;
	index: number;
	page: number;
	page_num: number;
	left: number;
	top: number;
};

function createSearchDragPreviewHost(
	srcEl: HTMLElement,
	clientX: number,
	clientY: number,
	mountRoot: HTMLElement | null
): { host: HTMLElement; offsetX: number; offsetY: number } {
	const rect = srcEl.getBoundingClientRect();
	const tileEl = srcEl.querySelector('.lp-app-tile');
	if (!(tileEl instanceof HTMLElement)) {
		throw new Error('Launchpad drag preview: missing .lp-app-tile');
	}
	const clone = tileEl.cloneNode(true) as HTMLElement;
	clone.querySelector('.lp-app-tile__drag-mask')?.remove();
	clone.querySelector('.delete_launch')?.remove();

	const shell = document.createElement('div');
	shell.className = 'lp-drag-preview-host';
	shell.style.setProperty('--lp-preview-left', `${rect.left}px`);
	shell.style.setProperty('--lp-preview-top', `${rect.top}px`);
	shell.style.setProperty('--lp-preview-width', `${rect.width}px`);
	shell.style.setProperty('--lp-preview-height', `${rect.height}px`);

	shell.appendChild(clone);

	const mount = mountRoot ?? document.body;
	mount.appendChild(shell);

	return {
		host: shell,
		offsetX: Math.max(0, Math.round(clientX - rect.left)),
		offsetY: Math.max(0, Math.round(clientY - rect.top))
	};
}

export function useLaunchpadDragReorder(opts: {
	slide: Ref<number>;
	carousel: Ref<{ goTo?: (n: number) => void } | undefined>;
	carouselAnimated?: Ref<boolean>;
	launchpadMount?: Ref<HTMLElement | undefined>;
	appStore: ReturnType<typeof useApplicationStore>;
	exitSearchForDrag: (draggedAppId?: string) => void;
	applyPendingSlideRestoreAfterDrag?: () => number | null;
	clearPendingSlideAfterSearchDrag?: () => void;
	resumeCarouselAfterSearchDrag?: () => void;
	onEmitDrag: (id: string | null) => void;
}) {
	const {
		slide,
		carousel,
		carouselAnimated,
		launchpadMount,
		appStore,
		exitSearchForDrag,
		applyPendingSlideRestoreAfterDrag,
		clearPendingSlideAfterSearchDrag,
		resumeCarouselAfterSearchDrag,
		onEmitDrag
	} = opts;

	const dragLaunchAppId = ref<string | null>(null);
	const launchPadDragSlotHidden = ref(false);
	const isInDrag = ref(false);
	const lastDragFinishTime = ref(0);

	let launchPadDragInitRaf = 0;
	let launchPadSearchExitRaf = 0;
	let launchPadDragFramesApplied = false;
	let choose = -1;
	let preDragSnapshot: DragSnapshotItem[] | null = null;
	let dragStartGeneration = 0;
	let searchDragPreviewHost: HTMLElement | null = null;
	let pendingSearchExitOnDragInit = false;

	function disposeSearchDragPreviewHost() {
		if (searchDragPreviewHost?.isConnected) {
			searchDragPreviewHost.remove();
		}
		searchDragPreviewHost = null;
	}

	function getCarouselRootEl(): HTMLElement | null {
		const c = carousel.value as { $el?: HTMLElement } | HTMLElement | undefined;
		if (!c) return null;
		if (c instanceof HTMLElement) return c;
		const el = c.$el;
		return el instanceof HTMLElement ? el : null;
	}

	function getEdgeFlipScreenBounds(): {
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

	function getLaunchpadSlideSurfaceEl(slideIndex: number): HTMLElement | null {
		const root = getCarouselRootEl();
		if (!root) return null;
		return root.querySelector(
			`[data-lp-slide="${slideIndex}"]`
		) as HTMLElement | null;
	}

	const EDGE_FLIP_DWELL_MS = 520;
	const EDGE_FLIP_COOLDOWN_MS = 720;
	const DOT_FLIP_DWELL_MS = 520;
	let dotFlipTimer: ReturnType<typeof setTimeout> | null = null;
	let dotFlipPendingIndex: number | null = null;

	function clearDotFlipDwell() {
		if (dotFlipTimer != null) {
			clearTimeout(dotFlipTimer);
			dotFlipTimer = null;
		}
		dotFlipPendingIndex = null;
	}

	const EDGE_LEAVE_GRACE_MS = 200;
	let lastDragOverClientX = 0;
	let lastDragOverClientY = 0;
	let lastDragReorderTargetIndex: number | null = null;
	let edgeFlipDwellTimer: ReturnType<typeof setTimeout> | null = null;
	let edgeFlipPendingDir: -1 | 0 | 1 = 0;
	let lastEdgeFlipAt = 0;
	let edgeLeaveGraceTimer: ReturnType<typeof setTimeout> | null = null;
	let edgeRepeatAfterFlipTimer: ReturnType<typeof setTimeout> | null = null;

	function clearEdgeFlipRepeatProbe() {
		if (edgeRepeatAfterFlipTimer != null) {
			clearTimeout(edgeRepeatAfterFlipTimer);
			edgeRepeatAfterFlipTimer = null;
		}
	}

	function scheduleEdgeFlipRepeatProbe() {
		clearEdgeFlipRepeatProbe();
		edgeRepeatAfterFlipTimer = setTimeout(() => {
			edgeRepeatAfterFlipTimer = null;
			if (!dragLaunchAppId.value) return;
			void nextTick(() => {
				if (!dragLaunchAppId.value) return;
				maybeFlipPageFromDragOver(lastDragOverClientX);
			});
		}, EDGE_FLIP_COOLDOWN_MS);
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

	function clearEdgeFlipDwell() {
		cancelEdgeLeaveGrace();
		if (edgeFlipDwellTimer != null) {
			clearTimeout(edgeFlipDwellTimer);
			edgeFlipDwellTimer = null;
		}
		edgeFlipPendingDir = 0;
	}

	function goto(value: number) {
		carousel.value?.goTo?.(value);
	}

	function isLaunchpadSlideMounted(slideIndex: number): boolean {
		if (slide.value !== slideIndex) return false;
		const surface = getLaunchpadSlideSurfaceEl(slideIndex);
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
					goto(target);
					void nextTick(() => trySync());
				});
				return;
			}

			slide.value = target;
			goto(target);

			const mounted = isLaunchpadSlideMounted(target);
			if (!mounted && attempts < 32) {
				void nextTick(() => trySync());
				return;
			}
		};

		void nextTick(() => {
			void nextTick(() => trySync());
		});
	}

	function maybeFlipPageFromDragOver(clientX: number) {
		lastDragOverClientX = clientX;
		if (!dragLaunchAppId.value) {
			clearEdgeFlipRepeatProbe();
			clearEdgeFlipDwell();
			return;
		}
		if (appStore.launchPadApps.length <= 1) {
			clearEdgeFlipRepeatProbe();
			clearEdgeFlipDwell();
			return;
		}

		const bounds = getEdgeFlipScreenBounds();
		const edgePx = Math.min(56, bounds.width * 0.04);

		let dir: -1 | 0 | 1 = 0;
		if (clientX <= bounds.left + edgePx && slide.value > 0) dir = -1;
		else if (
			clientX >= bounds.right - edgePx &&
			slide.value < appStore.launchPadApps.length - 1
		)
			dir = 1;

		if (dir === 0) {
			if (!edgeFlipDwellTimer) {
				scheduleEdgeLeaveGraceClear();
			}
			return;
		}

		cancelEdgeLeaveGrace();

		if (Date.now() - lastEdgeFlipAt < EDGE_FLIP_COOLDOWN_MS) {
			return;
		}

		if (edgeFlipPendingDir !== dir) {
			clearEdgeFlipRepeatProbe();
			clearEdgeFlipDwell();
			edgeFlipPendingDir = dir;
			const scheduledDir = dir;
			edgeFlipDwellTimer = setTimeout(() => {
				edgeFlipDwellTimer = null;
				edgeFlipPendingDir = 0;
				if (!dragLaunchAppId.value) return;
				if (Date.now() - lastEdgeFlipAt < EDGE_FLIP_COOLDOWN_MS) return;

				const b = getEdgeFlipScreenBounds();
				const ep = Math.min(56, b.width * 0.04);
				const x = lastDragOverClientX;
				let ok: -1 | 0 | 1 = 0;
				if (x <= b.left + ep && slide.value > 0) ok = -1;
				else if (
					x >= b.right - ep &&
					slide.value < appStore.launchPadApps.length - 1
				)
					ok = 1;
				if (ok === scheduledDir) {
					goto(slide.value + scheduledDir);
					lastEdgeFlipAt = Date.now();
					scheduleEdgeFlipRepeatProbe();
				}
			}, EDGE_FLIP_DWELL_MS);
		}
	}

	function onDocumentDragOverForEdge(e: Event) {
		if (!dragLaunchAppId.value) return;
		const de = e as DragEvent;
		de.preventDefault();
		try {
			if (de.dataTransfer) de.dataTransfer.dropEffect = 'move';
		} catch {
			void 0;
		}
		lastDragOverClientY = de.clientY;
		maybeFlipPageFromDragOver(de.clientX);
		maybeApplyReorderFromDragPointer(de.clientX, de.clientY);
	}

	function pointerToTargetGlobalIndex(
		clientX: number,
		clientY: number
	): number | null {
		const surface = getLaunchpadSlideSurfaceEl(slide.value);
		if (!surface) {
			return null;
		}
		const rect = surface.getBoundingClientRect();
		const margin = 12;
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

		const gridX = appStore.DESKTOP_APP_X_NUM;
		const gridY = appStore.DESKTOP_APP_Y_NUM;
		const perPage = gridX * gridY;
		const n = appStore.desktopApps.length;
		if (n === 0) return null;

		const SIZE = appStore.DESKTOP_APP_SIZE;
		const pitchX = SIZE + appStore.DESKTOP_APP_X_GAP;
		const pitchY = SIZE + appStore.DESKTOP_APP_Y_GAP;
		const START_X = appStore.DOCKER_APP_START_X_GAP;
		const START_Y = appStore.DOCKER_APP_START_Y_GAP;

		let col = Math.round((lx - START_X - SIZE / 2) / pitchX);
		let row = Math.round((ly - START_Y - SIZE / 2) / pitchY);
		col = Math.max(0, Math.min(gridX - 1, col));
		row = Math.max(0, Math.min(gridY - 1, row));

		const slot = row * gridX + col;
		const raw = slide.value * perPage + slot;
		return Math.max(0, Math.min(n - 1, raw));
	}

	const getAppIndexInDesktop = (id: string) =>
		appStore.desktopApps.findIndex((a) => a.id === id);

	function pushLaunchLayoutFromDesktopApps(): void {
		// The layout is a single shared linear order; emit { key, order } from
		// the current desktopApps sequence. Each device reflows it onto its own
		// grid at render time, so we don't encode page/slot here.
		const cur = appStore.layoutBySurface.launchpad;
		const sorted = [...appStore.desktopApps].sort((a, b) => a.index - b.index);
		const items: LaunchpadItem[] = [];
		const seen = new Set<string>();
		for (let i = 0; i < sorted.length; i++) {
			const app = sorted[i] as { id: string };
			const raw = app.id;
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

	function animate(sources: number[], targets: DesktopPosition[]) {
		if (sources.length === 0 || sources.length !== targets.length) return;

		for (let i = 0; i < sources.length; i++) {
			const s = sources[i];
			const t = targets[i];
			appStore.desktopApps[s].index = t.index;
			appStore.desktopApps[s].page = t.page;
			appStore.desktopApps[s].page_num = t.page_num;
			appStore.desktopApps[s].left = t.left;
			appStore.desktopApps[s].top = t.top;
		}

		rebuildLaunchPadApps();
	}

	function checkMoveToIndex(toIndex: number) {
		if (choose < 0 || choose >= appStore.desktopApps.length) return;

		const fromIndex = appStore.desktopApps[choose].index;
		if (fromIndex === toIndex) return;

		const sources: number[] = [];
		const targets: DesktopPosition[] = [];

		if (fromIndex < toIndex) {
			for (let i = 0; i < appStore.desktopApps.length; i++) {
				if (i == choose) {
					continue;
				}
				const idx = appStore.desktopApps[i].index;
				if (idx > fromIndex && idx <= toIndex) {
					sources.push(i);
					targets.push(appStore.get_desktop_position(idx - 1));
				}
			}
		} else {
			for (let i = 0; i < appStore.desktopApps.length; i++) {
				if (i == choose) {
					continue;
				}
				const idx = appStore.desktopApps[i].index;
				if (idx < fromIndex && idx >= toIndex) {
					sources.push(i);
					targets.push(appStore.get_desktop_position(idx + 1));
				}
			}
		}

		sources.push(choose);
		targets.push(appStore.get_desktop_position(toIndex));

		if (targets.length === 0) {
			return;
		}

		animate(sources, targets);
	}

	function maybeApplyReorderFromDragPointer(clientX: number, clientY: number) {
		if (!dragLaunchAppId.value || choose < 0) return;
		if (appStore.launchpadSearchQuery.length > 0) return;

		const el = document.elementFromPoint(clientX, clientY);
		let cur: HTMLElement | null = el as HTMLElement | null;
		while (cur && (!cur.id || !cur.id.startsWith('bdesk:'))) {
			cur = cur.parentElement;
		}
		if (cur && cur.id && cur.id !== dragLaunchAppId.value) {
			const hitIdx = getAppIndexInDesktop(cur.id);
			if (hitIdx >= 0) {
				const toIdx = appStore.desktopApps[hitIdx].index;
				if (lastDragReorderTargetIndex === toIdx) return;
				lastDragReorderTargetIndex = toIdx;
				checkMoveToIndex(toIdx);
				return;
			}
		}

		const fromGrid = pointerToTargetGlobalIndex(clientX, clientY);
		if (fromGrid !== null) {
			if (lastDragReorderTargetIndex === fromGrid) return;
			lastDragReorderTargetIndex = fromGrid;
			checkMoveToIndex(fromGrid);
			return;
		}

		lastDragReorderTargetIndex = null;
	}

	function installEdgeDragFlipListeners() {
		removeEdgeDragFlipListeners();
		document.addEventListener('dragover', onDocumentDragOverForEdge, {
			capture: true
		});
	}

	function removeEdgeDragFlipListeners() {
		clearEdgeFlipRepeatProbe();
		document.removeEventListener('dragover', onDocumentDragOverForEdge, {
			capture: true
		});
	}

	function applySearchExitAfterDragReady(
		startId: string,
		dragStartGen: number
	) {
		launchPadSearchExitRaf = requestAnimationFrame(() => {
			launchPadSearchExitRaf = 0;
			if (dragStartGen !== dragStartGeneration) return;
			if (!dragLaunchAppId.value) return;

			exitSearchForDrag(startId);
			pendingSearchExitOnDragInit = false;

			preDragSnapshot = appStore.desktopApps.map((a) => ({
				id: a.id,
				index: a.index,
				page: a.page,
				page_num: a.page_num,
				left: a.left,
				top: a.top
			}));

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

	function checkMove(id: string) {
		if (appStore.launchpadSearchQuery.length > 0) return;

		const dragIndex = getAppIndexInDesktop(id);

		if (choose == dragIndex || dragIndex == -1) {
			return;
		}

		const toIndex = appStore.desktopApps[dragIndex].index;
		checkMoveToIndex(toIndex);
	}

	const onDragStart = (e: DragEvent) => {
		const startId =
			(e.currentTarget as HTMLElement | null)?.id ||
			(e.target as HTMLElement)?.id ||
			'';

		const dt = e.dataTransfer;
		if (!dt) return;

		const wasSearching = appStore.launchpadSearchQuery.length > 0;
		const srcEl =
			e.currentTarget instanceof HTMLElement ? e.currentTarget : null;

		dt.setData('text', startId);
		dt.dropEffect = 'move';
		dt.effectAllowed = 'move';

		if (wasSearching && srcEl) {
			disposeSearchDragPreviewHost();
			try {
				const { host, offsetX, offsetY } = createSearchDragPreviewHost(
					srcEl,
					e.clientX,
					e.clientY,
					launchpadMount?.value ?? null
				);
				dt.setDragImage(host, offsetX, offsetY);
				searchDragPreviewHost = host;
				requestAnimationFrame(() => {
					if (searchDragPreviewHost === host) {
						host.style.left = '-99999px';
					}
				});
			} catch {
				disposeSearchDragPreviewHost();
			}
		}

		pendingSearchExitOnDragInit = wasSearching;

		launchPadDragFramesApplied = false;
		preDragSnapshot = null;
		if (launchPadDragInitRaf) {
			cancelAnimationFrame(launchPadDragInitRaf);
			launchPadDragInitRaf = 0;
		}

		const dragStartGen = ++dragStartGeneration;
		const deferSearchExit = pendingSearchExitOnDragInit;
		void nextTick(() => {
			if (dragStartGen !== dragStartGeneration) return;
			void nextTick(() => {
				if (dragStartGen !== dragStartGeneration) return;
				launchPadDragInitRaf = requestAnimationFrame(() => {
					launchPadDragInitRaf = 0;
					if (dragStartGen !== dragStartGeneration) return;

					if (!deferSearchExit) {
						preDragSnapshot = appStore.desktopApps.map((a) => ({
							id: a.id,
							index: a.index,
							page: a.page,
							page_num: a.page_num,
							left: a.left,
							top: a.top
						}));
					}

					isInDrag.value = true;
					choose = getAppIndexInDesktop(startId);

					dragLaunchAppId.value = startId;
					lastDragReorderTargetIndex = null;
					launchPadDragSlotHidden.value = false;
					window.setTimeout(() => {
						launchPadDragSlotHidden.value = true;
					}, 10);

					onEmitDrag(startId);

					installEdgeDragFlipListeners();
					launchPadDragFramesApplied = true;

					if (deferSearchExit) {
						applySearchExitAfterDragReady(startId, dragStartGen);
					} else {
						const restoredSlide = applyPendingSlideRestoreAfterDrag?.();
						if (restoredSlide != null) {
							goto(restoredSlide);
						}
					}
				});
			});
		});
	};

	const onDragEnd = async (e: DragEvent) => {
		disposeSearchDragPreviewHost();
		resumeCarouselAfterSearchDrag?.();
		dragStartGeneration++;
		if (launchPadDragInitRaf) {
			cancelAnimationFrame(launchPadDragInitRaf);
			launchPadDragInitRaf = 0;
		}
		if (launchPadSearchExitRaf) {
			cancelAnimationFrame(launchPadSearchExitRaf);
			launchPadSearchExitRaf = 0;
		}
		if (!launchPadDragFramesApplied) {
			clearPendingSlideAfterSearchDrag?.();
			pendingSearchExitOnDragInit = false;
		} else if (pendingSearchExitOnDragInit) {
			exitSearchForDrag(dragLaunchAppId.value ?? undefined);
			pendingSearchExitOnDragInit = false;
			if (!preDragSnapshot) {
				preDragSnapshot = appStore.desktopApps.map((a) => ({
					id: a.id,
					index: a.index,
					page: a.page,
					page_num: a.page_num,
					left: a.left,
					top: a.top
				}));
			}
		}

		const incomplete = !launchPadDragFramesApplied;

		removeEdgeDragFlipListeners();
		clearEdgeFlipDwell();
		clearDotFlipDwell();

		isInDrag.value = false;
		lastDragFinishTime.value = new Date().getTime();

		onEmitDrag(null);

		const cancelled =
			incomplete ||
			!e ||
			!e.dataTransfer ||
			e.dataTransfer.dropEffect === 'none';

		if (cancelled && preDragSnapshot) {
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
		} else if (!incomplete) {
			rebuildLaunchPadApps();
			const apps = appStore.desktopApps
				.map((a) => ({ id: a.id, index: a.index }))
				.sort((x, y) => x.index - y.index);
			localStorage.setItem(localStoreKeys.LAUNCH_PAD, JSON.stringify(apps));

			// Mirror drag-end into the cross-device launch layout. The store's
			// debounced setter collapses bursts into one PUT.
			pushLaunchLayoutFromDesktopApps();
		}

		preDragSnapshot = null;
		launchPadDragSlotHidden.value = false;
		dragLaunchAppId.value = null;
		lastDragReorderTargetIndex = null;
		launchPadDragFramesApplied = false;
	};

	const onDragEnter = async (id: string, e: any) => {
		e.preventDefault();

		if (!dragLaunchAppId.value) {
			return;
		}

		checkMove(id);
	};

	watch(slide, () => {
		if (!dragLaunchAppId.value) return;
		void nextTick(() => {
			requestAnimationFrame(() => {
				if (!dragLaunchAppId.value) return;
				maybeApplyReorderFromDragPointer(
					lastDragOverClientX,
					lastDragOverClientY
				);
			});
		});
	});

	const onDotDragEnter = (index: number) => {
		if (!dragLaunchAppId.value) return;
		if (slide.value === index) return;
		if (dotFlipPendingIndex === index) return;

		if (dotFlipTimer != null) clearTimeout(dotFlipTimer);
		dotFlipPendingIndex = index;
		dotFlipTimer = setTimeout(() => {
			dotFlipTimer = null;
			dotFlipPendingIndex = null;
			if (!dragLaunchAppId.value) return;
			if (slide.value === index) return;
			goto(index);
		}, DOT_FLIP_DWELL_MS);
	};

	const onDotDragLeave = () => {
		clearDotFlipDwell();
	};

	function dispose() {
		disposeSearchDragPreviewHost();
		removeEdgeDragFlipListeners();
		clearEdgeFlipDwell();
		clearDotFlipDwell();
		dragStartGeneration++;
		if (launchPadDragInitRaf) {
			cancelAnimationFrame(launchPadDragInitRaf);
			launchPadDragInitRaf = 0;
		}
		if (launchPadSearchExitRaf) {
			cancelAnimationFrame(launchPadSearchExitRaf);
			launchPadSearchExitRaf = 0;
		}
	}

	return {
		dragLaunchAppId,
		launchPadDragSlotHidden,
		isInDrag,
		onDragStart,
		onDragEnd,
		onDragEnter,
		onDotDragEnter,
		onDotDragLeave,
		goto,
		lastDragFinishTime,
		dispose
	};
}
