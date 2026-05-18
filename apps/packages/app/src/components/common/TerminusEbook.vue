<template>
	<div class="reader">
		<div class="toolbar">
			<div class="toolbar__left">
				<button
					v-if="showInternalToc"
					type="button"
					class="btn"
					@click="toggleToc"
				>
					{{ showToc ? t('Close Table of Contents') : t('Table of Contents') }}
				</button>
				<button type="button" class="btn" @click="prev">
					{{ t('Previous Page') }}
				</button>
				<button type="button" class="btn" @click="next">
					{{ t('Next Page') }}
				</button>

				<select
					v-if="showReadingModeSwitcher"
					class="select"
					v-model="readingMode"
				>
					<option value="scroll">{{ t('Scroll') }}</option>
					<option value="spread">{{ t('Page Turn') }}</option>
				</select>
			</div>

			<div class="toolbar__center">
				<div class="progress" v-if="showInternalProgress && progress != null">
					<input
						type="range"
						min="0"
						max="100"
						step="0.1"
						:value="progress * 100"
						@change="onProgressSliderChange"
					/>
					<span class="progress__text">{{ Math.round(progress * 100) }}%</span>
				</div>
			</div>

			<div class="toolbar__right">
				<select class="select" v-model="theme" @change="applyTheme">
					<option value="light">{{ t('Light') }}</option>
					<option value="sepia">{{ t('Sepia') }}</option>
					<option value="dark">{{ t('Dark') }}</option>
				</select>
				<button type="button" class="btn" @click="fontScaleDown">A-</button>
				<button type="button" class="btn" @click="fontScaleUp">A+</button>
				<button type="button" class="btn" @click="resetTypography">
					{{ t('Reset') }}
				</button>
			</div>
		</div>

		<div class="toolbar__content">
			<aside v-if="showInternalToc && showToc" class="toc">
				<div class="toc__title">{{ t('Table of Contents') }}</div>
				<div v-if="tocLoading" class="toc__hint">{{ t('Loading...') }}</div>
				<div v-else-if="toc.length === 0" class="toc__hint">
					{{ t('No table of contents available') }}
				</div>
				<ul v-else class="toc__list">
					<li
						v-for="item in toc"
						:key="item.id || item.href || item.label"
						class="toc__item"
					>
						<button type="button" class="toc__btn" @click="goToTocItem(item)">
							{{ item.label }}
						</button>
					</li>
				</ul>
			</aside>

			<div class="viewerWrap">
				<div
					ref="viewerRef"
					class="viewer"
					:class="{ 'viewer--ready': firstRendered }"
				/>

				<button
					v-if="readingMode === 'spread'"
					type="button"
					class="pageZone pageZone--left"
					:aria-label="t('Previous Page')"
					@click="prev"
				/>
				<button
					v-if="readingMode === 'spread'"
					type="button"
					class="pageZone pageZone--right"
					:aria-label="t('Next Page')"
					@click="next"
				/>

				<div v-if="loading" class="overlay">
					<div class="overlay__card">{{ t('Loading...') }}</div>
				</div>
				<div v-else-if="error" class="overlay overlay--error">
					<div class="overlay__card">
						<div class="overlay__title">{{ t('Loading Failed') }}</div>
						<div class="overlay__msg">{{ error }}</div>
						<button type="button" class="btn" @click="reload">
							{{ t('Retry') }}
						</button>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import {
	computed,
	nextTick,
	onBeforeUnmount,
	onMounted,
	ref,
	watch
} from 'vue';
import { useReadingProgressStore } from 'src/stores/rss-reading-progress';
import { useUserStore } from 'src/stores/user';
import { useI18n } from 'vue-i18n';

type TocItem = {
	id?: string;
	label: string;
	href?: string;
	subitems?: TocItem[];
};

type ArticleTopic = {
	text: string;
	id: string;
	level: number;
	jump: boolean;
};

const props = withDefaults(
	defineProps<{
		src: string;
		playedTime?: string; // Compatible with legacy field: usually CFI
		initialCfi?: string;
		accessToken?: string;
		requestHeaders?: Record<string, string>;
		manager?: 'continuous' | 'default';
		flow?: 'scrolled' | 'paginated';
		allowScriptedContent?: boolean;
		generateLocations?: boolean;
		locationsChars?: number;
		cacheLocations?: boolean;
		cacheKey?: string;
		initialTheme?: 'light' | 'sepia' | 'dark';
		initialFontScale?: number; // 1 = 100%
		showInternalProgress?: boolean;
		showInternalToc?: boolean;
		initialReadingMode?: 'scroll' | 'spread';
		showReadingModeSwitcher?: boolean;
	}>(),
	{
		allowScriptedContent: true,
		generateLocations: true,
		locationsChars: 1600,
		cacheLocations: true,
		initialTheme: 'light',
		initialFontScale: 1,
		showInternalProgress: true,
		showInternalToc: true,
		initialReadingMode: 'scroll',
		showReadingModeSwitcher: true
	}
);

const emit = defineEmits<{
	(e: 'ready'): void;
	(e: 'error', message: string): void;
	(e: 'toc', toc: TocItem[]): void;
	(e: 'articleTopics', topics: ArticleTopic[]): void;
	(e: 'relocated', payload: any): void;
	(e: 'progress', percentage: number | null): void;
	(e: 'locationCfi', cfi: string | null): void;
	(e: 'totalLocations', total: number | null): void;
}>();

const { t } = useI18n();
const viewerRef = ref<HTMLElement | null>(null);
const book = ref<any | null>(null);
const rendition = ref<any | null>(null);

const loading = ref(false);
const error = ref<string | null>(null);
const firstRendered = ref(false);

const tocLoading = ref(false);
const toc = ref<TocItem[]>([]);
const showToc = ref(false);

const theme = ref<'light' | 'sepia' | 'dark'>(props.initialTheme);
const fontScale = ref<number>(props.initialFontScale);
const lineHeight = ref<number>(1.6);

const progress = ref<number | null>(null);
const totalLocations = ref<number | null>(null);
const articleTopics = ref<ArticleTopic[]>([]);

const currentCfi = ref<string | null>(null);

const readingProgressStore = useReadingProgressStore?.();
const userStore = useUserStore?.();

let resizeObserver: ResizeObserver | null = null;
let initSeq = 0;

const READING_MODE_KEY = 'epubjs:readingMode';

function getSavedReadingMode(): 'scroll' | 'spread' {
	try {
		const s = localStorage.getItem(READING_MODE_KEY);
		if (s === 'scroll' || s === 'spread') return s;
	} catch {
		// ignore
	}
	return props.initialReadingMode;
}

const readingMode = ref<'scroll' | 'spread'>(getSavedReadingMode());

const targetInitialCfi = computed(
	() => currentCfi.value || props.playedTime || props.initialCfi || null
);
const locationsStorageKey = computed(() => {
	const key = props.cacheKey?.trim();
	if (key) return `epubjs:locations:${key}`;
	return `epubjs:locations:${props.src}`;
});

function toggleToc() {
	showToc.value = !showToc.value;
}

function prev() {
	if (!rendition.value) return;
	rendition.value.prev();
}

function next() {
	if (!rendition.value) return;
	rendition.value.next();
}

function fontScaleUp() {
	fontScale.value = Math.min(2, +(fontScale.value + 0.1).toFixed(2));
	applyTypography();
}

function fontScaleDown() {
	fontScale.value = Math.max(0.6, +(fontScale.value - 0.1).toFixed(2));
	applyTypography();
}

function resetTypography() {
	fontScale.value = props.initialFontScale;
	lineHeight.value = 1.6;
	applyTypography();
}

const THEME_BODY_STYLES: Record<
	string,
	Record<string, Record<string, string>>
> = {
	light: {
		body: { background: '#ffffff', color: '#1f2328' },
		a: { color: '#0969da' }
	},
	sepia: {
		body: { background: '#fbf1d0', color: '#3a2f1b' },
		a: { color: '#8a4b08' }
	},
	dark: {
		body: { background: '#0d1117', color: '#e6edf3' },
		a: { color: '#2f81f7' }
	}
};

//
// theme color select restore issue caused by high-level dark CSS styles (need reload CSS for direct style setting)
// https://github.com/futurepress/epub.js/issues/1101
// https://github.com/futurepress/epub.js/issues/1208#issuecomment-1724915756
// function applyTheme() {
// 	if (!rendition.value) return;
// 	try {
// 		const themesApi = rendition.value.themes;
// 		themesApi.select(theme.value);
// 		const styles = THEME_BODY_STYLES[theme.value];
// 		if (styles) {
// 			themesApi.default(styles);
// 		}
// 	} catch (e) {
// 		console.log(e);
// 	}
// }

function applyTheme() {
	if (!rendition.value) return;
	try {
		const style = THEME_BODY_STYLES[theme.value];
		rendition.value.themes.override('background', style.body.background, true);
		rendition.value.themes.override('color', style.body.color, true);
		rendition.value.themes.default({
			a: {
				color: `${style.a.color} !important`
			}
		});
	} catch (e) {
		console.log(e);
	}
}

function applyTypography() {
	if (!rendition.value) return;
	try {
		rendition.value.themes.fontSize(`${Math.round(fontScale.value * 100)}%`);
	} catch {
		// ignore
	}
}

async function goToTocItem(item: TocItem) {
	if (!rendition.value) return;
	const href = item.href;
	if (!href) return;
	try {
		await rendition.value.display(href);
		showToc.value = false;
	} catch (e: any) {
		const msg = e?.message || 'goToTocItem failure';
		error.value = msg;
		emit('error', msg);
	}
}

async function onProgressSliderChange(ev: Event) {
	if (!book.value || !rendition.value || !book.value.locations) return;
	const el = ev.target as HTMLInputElement;
	const pct = Number(el.value) / 100;
	try {
		const cfi = book.value.locations.cfiFromPercentage(pct);
		await rendition.value.display(cfi);
	} catch (e: any) {
		const msg = e?.message || 'progress slider change failure';
		error.value = msg;
		emit('error', msg);
	}
}

function cleanup() {
	loading.value = false;
	error.value = null;
	firstRendered.value = false;
	progress.value = null;
	totalLocations.value = null;

	if (resizeObserver) {
		try {
			resizeObserver.disconnect();
		} catch {
			// ignore
		}
		resizeObserver = null;
	}

	if (rendition.value) {
		try {
			rendition.value.destroy?.();
		} catch {
			// ignore
		}
		rendition.value = null;
	}

	if (book.value) {
		try {
			book.value.destroy?.();
		} catch {
			// ignore
		}
		book.value = null;
	}
}

function registerThemes(r: any) {
	try {
		r.themes.register('light', {
			body: { background: '#ffffff', color: '#1f2328' },
			a: { color: '#0969da' }
		});
		r.themes.register('sepia', {
			body: { background: '#fbf1d0', color: '#3a2f1b' },
			a: { color: '#8a4b08' }
		});
		r.themes.register('dark', {
			body: { background: '#0d1117', color: '#e6edf3' },
			a: { color: '#2f81f7' }
		});
	} catch {
		// ignore
	}
}

function setupResizeObserver(r: any) {
	if (!viewerRef.value) return;
	let raf = 0;
	resizeObserver = new ResizeObserver(() => {
		if (raf) cancelAnimationFrame(raf);
		raf = requestAnimationFrame(() => {
			try {
				r.resize?.();
			} catch {
				// ignore
			}
		});
	});
	resizeObserver.observe(viewerRef.value);
}

function applyReadingMode(r: any) {
	try {
		if (readingMode.value === 'spread') {
			r.flow?.('paginated');
			r.spread?.('always');
		} else {
			r.flow?.('scrolled');
			r.spread?.('none');
		}
	} catch {
		// ignore
	}
}

async function ensureContainerSized(el: HTMLElement, seq: number) {
	// Wait until the container has a stable size to avoid blank spaces or flickering caused by width/height=0 during initialization.
	for (let i = 0; i < 30; i++) {
		if (seq !== initSeq) return false;
		const rect = el.getBoundingClientRect();
		if (rect.width > 10 && rect.height > 10) return true;
		await new Promise((r) => requestAnimationFrame(() => r(null)));
	}
	return true;
}

async function loadLocationsIfNeeded(seq: number) {
	if (!book.value?.locations || !props.generateLocations) {
		emit('totalLocations', null);
		return;
	}

	if (props.cacheLocations) {
		try {
			const saved = localStorage.getItem(locationsStorageKey.value);
			if (saved) {
				book.value.locations.load(saved);
			}
		} catch {
			// ignore
		}
	}

	if (seq !== initSeq) return;

	// Generate only when there is no cache (generation may be slow and cause a momentary blank screen).
	if (!book.value.locations.length?.() || book.value.locations.length() === 0) {
		await book.value.locations.generate(props.locationsChars);
		if (seq !== initSeq) return;
		if (props.cacheLocations) {
			try {
				const data = book.value.locations.save();
				localStorage.setItem(locationsStorageKey.value, data);
			} catch {
				// ignore
			}
		}
	}

	const total = Number(book.value.locations.length?.() || 0) || null;
	totalLocations.value = total;
	emit('totalLocations', total);

	try {
		readingProgressStore.setTotalProgress(1, total);
	} catch {
		// ignore
	}
}

function buildArticleTopics(
	items: TocItem[],
	level = 0,
	acc: ArticleTopic[] = []
): ArticleTopic[] {
	for (const item of items || []) {
		const label = item.label || '';
		const href = item.href || '';
		let id = '';

		if (href.includes('#')) {
			id = href.split('#')[1] || href;
		} else {
			id = href || label;
		}

		acc.push({
			text: label,
			id,
			level,
			jump: !!href
		});

		if (item.subitems?.length) {
			buildArticleTopics(item.subitems, level + 1, acc);
		}
	}

	return acc;
}

async function init() {
	const seq = ++initSeq;
	cleanup();
	loading.value = true;
	tocLoading.value = true;
	toc.value = [];

	await nextTick();
	const container = viewerRef.value;
	if (!container) {
		loading.value = false;
		return;
	}
	await ensureContainerSized(container, seq);
	if (seq !== initSeq) return;

	try {
		const ePub = (await import('epubjs')).default;
		const headers: Record<string, string> = { ...(props.requestHeaders || {}) };

		const token =
			props.accessToken ||
			(userStore && (userStore as any).current_user?.access_token) ||
			'';
		if (token) headers['X-Authorization'] = token;

		book.value = ePub(
			props.src,
			Object.keys(headers).length ? { requestHeaders: headers } : undefined
		);

		const renderOptions: any = {
			manager: props.manager ?? 'continuous',
			flow: props.flow ?? 'scrolled',
			width: '100%',
			height: '100%',
			allowScriptedContent: props.allowScriptedContent
		};

		rendition.value = book.value.renderTo(container, renderOptions);

		registerThemes(rendition.value);
		applyTheme();
		applyTypography();
		applyReadingMode(rendition.value);
		setupResizeObserver(rendition.value);

		rendition.value.on('rendered', () => {
			firstRendered.value = true;
		});

		rendition.value.on('displayError', (e: any) => {
			const msg = e?.message || 'displayError';
			error.value = msg;
			emit('error', msg);
		});

		rendition.value.on('relocated', (location: any) => {
			emit('relocated', location);
			const cfi = location?.start?.cfi || null;
			emit('locationCfi', cfi);
			currentCfi.value = cfi;

			if (book.value?.locations && cfi) {
				try {
					const pct = book.value.locations.percentageFromCfi(cfi);
					progress.value = typeof pct === 'number' ? pct : null;

					if (typeof pct === 'number') {
						readingProgressStore.updateProgress(pct);
					}
				} catch {
					progress.value = null;
				}
			}
			emit('progress', progress.value);
		});

		book.value.loaded.navigation
			.then((nav: any) => {
				const items: TocItem[] = nav?.toc || nav || [];
				toc.value = Array.isArray(items) ? items : [];
				const topics = buildArticleTopics(toc.value);
				articleTopics.value = topics;

				emit('toc', toc.value);
				emit('articleTopics', topics);
			})
			.catch(() => {
				toc.value = [];
			})
			.finally(() => {
				tocLoading.value = false;
			});

		book.value.ready
			.then(async () => {
				if (seq !== initSeq) return;
				await loadLocationsIfNeeded(seq);
			})
			.catch(() => {
				emit('totalLocations', null);
			});

		const toDisplay = targetInitialCfi.value;
		try {
			await rendition.value.display(toDisplay || undefined);
		} finally {
			loading.value = false;
			emit('ready');
		}
	} catch (e: any) {
		loading.value = false;
		const msg = e?.message || 'ePub load failure';
		error.value = msg;
		emit('error', msg);
	}
}

async function reload() {
	await init();
}

watch(
	() => props.src,
	() => {
		init();
	},
	{ immediate: false }
);

watch(
	() => props.accessToken,
	() => {
		// A change in the token usually means re-requesting resources; to avoid flickering during partial page rendering, reload the page directly.
		if (book.value) init();
	}
);

watch(
	() => readingMode.value,
	(val) => {
		try {
			localStorage.setItem(READING_MODE_KEY, val);
		} catch {
			// ignore
		}
		if (rendition.value) {
			applyReadingMode(rendition.value);
		}
	}
);

onMounted(() => {
	const saved = getSavedReadingMode();
	if (saved !== readingMode.value) {
		readingMode.value = saved;
	}
	init();
	window.addEventListener('keydown', onKeydown, { passive: true });
});

function onKeydown(e: KeyboardEvent) {
	if (e.key === 'ArrowLeft') prev();
	if (e.key === 'ArrowRight') next();
}

onBeforeUnmount(() => {
	window.removeEventListener('keydown', onKeydown);
	cleanup();
});
</script>

<style scoped>
.reader {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	background: #fff;
}

.toolbar {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 12px;
	padding: 10px 12px;
	border-bottom: 1px solid rgba(0, 0, 0, 0.08);
	background: rgba(255, 255, 255, 0.92);
	backdrop-filter: blur(8px);
}

.toolbar__left,
.toolbar__center,
.toolbar__right {
	display: flex;
	align-items: center;
	gap: 8px;
	min-width: 0;
}

.toolbar__center {
	flex: 1;
	justify-content: center;
}

.btn {
	border: 1px solid rgba(0, 0, 0, 0.12);
	background: #fff;
	color: #1f2328;
	border-radius: 8px;
	padding: 6px 10px;
	font-size: 12px;
	cursor: pointer;
}

.btn:hover {
	background: rgba(0, 0, 0, 0.03);
}

.select {
	border: 1px solid rgba(0, 0, 0, 0.12);
	background: #fff;
	border-radius: 8px;
	padding: 6px 8px;
	font-size: 12px;
}

.progress {
	display: flex;
	align-items: center;
	gap: 8px;
	min-width: 220px;
	max-width: 520px;
	width: 100%;
}

.progress input[type='range'] {
	width: 100%;
}

.progress__text {
	font-size: 12px;
	color: rgba(31, 35, 40, 0.75);
	min-width: 42px;
	text-align: right;
}

.toolbar__content {
	flex: 1;
	min-height: 0;
	display: flex;
}

.toc {
	width: 260px;
	border-right: 1px solid rgba(0, 0, 0, 0.08);
	background: rgba(250, 250, 250, 0.98);
	padding: 10px;
	overflow: auto;
}

.toc__title {
	font-weight: 600;
	font-size: 13px;
	margin-bottom: 8px;
}

.toc__hint {
	font-size: 12px;
	color: rgba(31, 35, 40, 0.7);
	padding: 8px 0;
}

.toc__list {
	list-style: none;
	padding: 0;
	margin: 0;
	display: flex;
	flex-direction: column;
	gap: 6px;
}

.toc__btn {
	width: 100%;
	text-align: left;
	border: 0;
	background: transparent;
	padding: 6px 8px;
	border-radius: 8px;
	cursor: pointer;
	font-size: 12px;
	color: #0969da;
}

.toc__btn:hover {
	background: rgba(9, 105, 218, 0.08);
}

.viewerWrap {
	flex: 1;
	min-width: 0;
	position: relative;
}

.viewer {
	width: 100%;
	height: 100%;
	opacity: 0;
	transition: opacity 120ms ease;
}

.viewer--ready {
	opacity: 1;
}

.pageZone {
	position: absolute;
	top: 0;
	bottom: 0;
	width: 10%;
	min-width: 32px;
	max-width: 56px;
	border: 0;
	padding: 0;
	margin: 0;
	background: transparent;
	cursor: pointer;
	z-index: 1;
	transition: background 140ms ease;
}

.pageZone--left {
	left: 0;
}

.pageZone--right {
	right: 0;
}

.pageZone:hover {
	background: radial-gradient(
		farthest-side at 50% 50%,
		rgba(0, 0, 0, 0.08),
		transparent 75%
	);
}

.overlay {
	position: absolute;
	inset: 0;
	display: flex;
	align-items: center;
	justify-content: center;
	background: rgba(255, 255, 255, 0.75);
}

.overlay--error {
	background: rgba(255, 255, 255, 0.92);
}

.overlay__card {
	max-width: 520px;
	width: calc(100% - 24px);
	border: 1px solid rgba(0, 0, 0, 0.12);
	border-radius: 12px;
	background: #fff;
	padding: 14px;
	box-shadow: 0 12px 30px rgba(0, 0, 0, 0.08);
	display: flex;
	flex-direction: column;
	gap: 8px;
}

.overlay__title {
	font-weight: 600;
}

.overlay__msg {
	font-size: 12px;
	color: rgba(31, 35, 40, 0.75);
	word-break: break-word;
}
</style>
