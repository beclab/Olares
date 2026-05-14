<template>
	<div
		ref="containerRef"
		class="thumbnail-wrapper cursor-pointer"
		:class="{ active: active }"
		@click="$emit('click')"
	>
		<!-- Thumbnail with border -->
		<div class="thumbnail-content" :class="active ? 'selected' : ''">
			<!-- Loaded: show thumbnail image -->
			<img
				v-if="thumbnailUrl"
				:src="thumbnailUrl"
				:alt="`Page ${page}`"
				class="thumbnail-image"
			/>
			<!-- Loading: show spinner -->
			<div v-else-if="isLoading" class="loading-inner">
				<div class="loading-spinner"></div>
			</div>
			<!-- Not loaded: show placeholder -->
			<div v-else class="placeholder-inner">
				<span class="placeholder-icon">📄</span>
			</div>
		</div>
		<!-- Page number outside border -->
		<div class="page-label" :class="{ active: active }">{{ page }}</div>
	</div>
</template>

<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { getThumbnail, ThumbnailOptions } from '../services/pdfThumbnail';

const props = defineProps<{
	pdfDocument: any;
	page: number;
	sourceKey: string;
	active?: boolean;
	width?: number;
}>();

defineEmits<{
	click: [];
}>();

const containerRef = ref<HTMLElement | null>(null);
const thumbnailUrl = ref<string>('');
const isLoading = ref(false);
const hasError = ref(false);
const isInViewport = ref(false);

let observer: IntersectionObserver | null = null;

const thumbnailOptions: ThumbnailOptions = {
	width: props.width || 120,
	quality: 0.6,
	format: 'image/jpeg'
};

async function loadThumbnail() {
	if (!props.pdfDocument) return;
	if (thumbnailUrl.value) return; // Already loaded
	if (isLoading.value) return; // Already loading

	isLoading.value = true;
	hasError.value = false;

	try {
		thumbnailUrl.value = await getThumbnail(
			props.pdfDocument,
			props.page,
			props.sourceKey,
			thumbnailOptions
		);
	} catch (error) {
		console.warn(`Failed to load thumbnail for page ${props.page}:`, error);
		hasError.value = true;
	} finally {
		isLoading.value = false;
	}
}

// Setup Intersection Observer for viewport detection
function setupObserver() {
	if (observer) return;
	if (!containerRef.value) return;

	observer = new IntersectionObserver(
		(entries) => {
			const entry = entries[0];
			if (entry) {
				isInViewport.value = entry.isIntersecting;
				// Load thumbnail when entering viewport
				if (entry.isIntersecting && !thumbnailUrl.value && props.pdfDocument) {
					loadThumbnail();
				}
			}
		},
		{
			root: null, // Use viewport
			rootMargin: '100px', // Preload when 100px away from viewport
			threshold: 0
		}
	);

	observer.observe(containerRef.value);
}

function cleanupObserver() {
	if (observer) {
		observer.disconnect();
		observer = null;
	}
}

// Load when pdfDocument becomes available and in viewport
watch(
	() => props.pdfDocument,
	() => {
		if (props.pdfDocument && isInViewport.value && !thumbnailUrl.value) {
			loadThumbnail();
		}
	}
);

// Scroll into view when becoming active
watch(
	() => props.active,
	(isActive) => {
		if (isActive && containerRef.value) {
			containerRef.value.scrollIntoView({
				behavior: 'smooth',
				block: 'center'
			});
		}
	}
);

onMounted(() => {
	setupObserver();
	// If already active on mount, scroll into view
	if (props.active && containerRef.value) {
		setTimeout(() => {
			containerRef.value?.scrollIntoView({
				behavior: 'smooth',
				block: 'center'
			});
		}, 300);
	}
});

onBeforeUnmount(() => {
	cleanupObserver();
});
</script>

<style scoped lang="scss">
.thumbnail-wrapper {
	margin-bottom: 12px;
	display: flex;
	flex-direction: column;
	align-items: center;
}

.thumbnail-content {
	width: 90px;
	min-height: 120px;
	padding: 4px;
	background: $background-3;
	border-radius: 8px;
	border: 2px solid $separator;
	transition: border-color 0.15s ease, background-color 0.15s ease;

	&.selected {
		background: rgba($orange-default, 0.1);
		border-color: $orange-default;
	}
}

.thumbnail-image {
	display: block;
	width: 100%;
	border-radius: 4px;
}

.placeholder-inner {
	width: 100%;
	height: 120px;
	display: flex;
	align-items: center;
	justify-content: center;
	background: linear-gradient(135deg, #f5f7fa 0%, #e8eaed 100%);
	border-radius: 4px;
}

.placeholder-icon {
	font-size: 24px;
	opacity: 0.5;
}

.loading-inner {
	width: 100%;
	height: 120px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.page-label {
	margin-top: 6px;
	font-size: 12px;
	color: $ink-2;
	font-weight: 500;
	text-align: center;

	&.active {
		color: $orange-default;
		font-weight: 600;
	}
}

.loading-spinner {
	width: 24px;
	height: 24px;
	border: 2px solid #e0e0e0;
	border-top-color: $orange-default;
	border-radius: 50%;
	animation: spin 0.8s linear infinite;
}

@keyframes spin {
	to {
		transform: rotate(360deg);
	}
}
</style>
