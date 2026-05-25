import {
	ref,
	computed,
	nextTick,
	onMounted,
	onBeforeUnmount,
	watch,
	type Ref
} from 'vue';
import { useApplicationStore } from 'src/stores/desktop/app';

const LAUNCHPAD_GRID_ROW_TRACK_PX = 91;

export function useMobileLaunchpad(
	containerEl: Ref<HTMLElement | null>,
	carousel: Ref<any>,
	slide: Ref<number>,
	searchInputFocused?: Ref<boolean>,
	keyboardOpen?: Ref<boolean>
) {
	const appStore = useApplicationStore();

	const slideGridStyle = computed(() => {
		const override = appStore.launchpadGridOverride;
		return {
			'--lp-cols': override?.x ?? 4,
			'--lp-rows': override?.y ?? 4,
			'--lp-row-track': `${LAUNCHPAD_GRID_ROW_TRACK_PX}px`
		};
	});

	let resizeDebounce: ReturnType<typeof setTimeout> | null = null;
	let resizeObserver: ResizeObserver | null = null;

	function setGrid(cols: number, rows: number): boolean {
		const cx = Math.max(3, Math.min(8, Math.round(cols)));
		const ry = Math.max(3, Math.min(10, Math.round(rows)));
		const prev = appStore.launchpadGridOverride;
		if (prev?.x === cx && prev?.y === ry) return false;
		appStore.launchpadGridOverride = { x: cx, y: ry };
		if (appStore.myApps.length) {
			const search = appStore.launchpadSearchQuery;
			if (search.length > 0) {
				void appStore.update_search_result(search);
			} else {
				appStore.relocate_application_place(appStore.myApps);
			}
		}
		return true;
	}

	function applyGridFromSize() {
		const el = containerEl.value;
		if (!el) return;
		if (keyboardOpen?.value) return;

		const padTop = 10;
		const paddingLeft = 12;
		const dotsH = 60;
		const rowGap = 30;
		const h = el.clientHeight;
		const w = el.clientWidth - paddingLeft * 2;
		if (h < 8 || w < 8) return;

		const stride = LAUNCHPAD_GRID_ROW_TRACK_PX + rowGap;
		const gridH = h - padTop - dotsH;

		let rows = Math.floor((gridH + rowGap) / stride);
		rows = Math.max(3, Math.min(10, rows));

		const minColW = 72;
		let cols = Math.floor(w / minColW);
		cols = Math.max(3, Math.min(8, cols));

		const gridChanged = setGrid(cols, rows);
		if (gridChanged) {
			if (searchInputFocused?.value) {
				return;
			}
			const pageCount = appStore.launchPadApps?.length ?? 0;
			const maxSlide = Math.max(0, pageCount - 1);
			if (slide.value > maxSlide) {
				slide.value = maxSlide;
			}
			const target = slide.value;
			nextTick(() => {
				carousel.value?.goTo?.(target);
			});
		}
	}

	function scheduleApplyGrid() {
		if (resizeDebounce) clearTimeout(resizeDebounce);
		resizeDebounce = setTimeout(() => {
			resizeDebounce = null;
			applyGridFromSize();
		}, 100);
	}

	function tryAttachResize() {
		const el = containerEl.value;
		if (!el) return;
		applyGridFromSize();
		if (!resizeObserver) {
			resizeObserver = new ResizeObserver(() => scheduleApplyGrid());
			resizeObserver.observe(el);
		}
	}

	onMounted(() => {
		nextTick(() => tryAttachResize());
	});

	watch(
		() => appStore.launchPadApps?.length ?? 0,
		(len) => {
			if (!len) {
				resizeObserver?.disconnect();
				resizeObserver = null;
				return;
			}
			nextTick(() => tryAttachResize());
		}
	);

	onBeforeUnmount(() => {
		if (resizeDebounce) clearTimeout(resizeDebounce);
		resizeObserver?.disconnect();
		resizeObserver = null;
	});

	return {
		slideGridStyle
	};
}
