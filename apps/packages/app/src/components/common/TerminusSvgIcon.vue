<template>
	<span class="terminus-svg-icon" aria-hidden="true" v-html="svgMarkup" />
</template>

<script lang="ts">
// Inline-SVG renderer used instead of `<img :src="require(...)">` for assets
// that intermittently fail to paint in the iOS/Android WebView (the tab bar
// icons). Loading the SVG markup once and injecting it into the DOM avoids the
// per-element image decode/scheme-load race entirely; a module-level cache
// makes every later mount synchronous so the icon can never go blank after the
// first successful load, and a failed load is retried on the next mount.
const svgCache = new Map<string, string>();
const inflight = new Map<string, Promise<string>>();

function loadSvg(url: string): Promise<string> {
	const cached = svgCache.get(url);
	if (cached !== undefined) {
		return Promise.resolve(cached);
	}
	let pending = inflight.get(url);
	if (!pending) {
		pending = fetch(url)
			.then((res) => (res.ok ? res.text() : ''))
			.then((text) => {
				inflight.delete(url);
				// Only cache real SVG markup; never cache an empty/failed result so
				// a later mount retries instead of permanently showing blank.
				if (text && text.includes('<svg')) {
					svgCache.set(url, text);
					return text;
				}
				return '';
			})
			.catch(() => {
				inflight.delete(url);
				return '';
			});
		inflight.set(url, pending);
	}
	return pending;
}
</script>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { getRequireImage } from '../../utils/imageUtils';

const props = defineProps<{ path: string }>();

const svgMarkup = ref('');

const apply = async (path: string) => {
	if (!path) {
		svgMarkup.value = '';
		return;
	}
	const url = getRequireImage(path);
	const cached = svgCache.get(url);
	if (cached !== undefined) {
		svgMarkup.value = cached;
		return;
	}
	const text = await loadSvg(url);
	// Guard against a prop change while the fetch was in flight.
	if (getRequireImage(props.path) === url) {
		svgMarkup.value = text;
	}
};

watch(() => props.path, apply, { immediate: true });
</script>

<style scoped lang="scss">
.terminus-svg-icon {
	display: inline-flex;
	line-height: 0;

	:deep(svg) {
		width: 100%;
		height: 100%;
		display: block;
	}
}
</style>
