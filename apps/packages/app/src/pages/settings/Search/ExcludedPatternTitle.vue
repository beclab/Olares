<template>
	<div class="excluded-pattern-title-root">
		<!-- Quasar tooltip treats the direct parent of q-tooltip as activator -->
		<div class="excluded-pattern-activator">
			<span ref="textRef" class="excluded-pattern-clamp">{{ text }}</span>
			<bt-tooltip
				v-if="showTooltip"
				class="text-body3 excluded-pattern-tooltip"
				:label="text"
				max-width="480px"
				align="start"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import BtTooltip from 'src/components/base/BtTooltip.vue';
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue';

const props = defineProps<{ text: string }>();

const textRef = ref<HTMLElement | null>(null);
const showTooltip = ref(false);

function updateOverflow() {
	const el = textRef.value;
	if (!el) {
		showTooltip.value = false;
		return;
	}
	showTooltip.value = el.scrollHeight > el.clientHeight + 1;
}

let ro: ResizeObserver | undefined;

onMounted(() => {
	nextTick(() => {
		updateOverflow();
		if (textRef.value) {
			ro = new ResizeObserver(() => updateOverflow());
			ro.observe(textRef.value);
		}
	});
});

watch(
	() => props.text,
	() => nextTick(updateOverflow)
);

onBeforeUnmount(() => {
	ro?.disconnect();
});
</script>

<style scoped lang="scss">
.excluded-pattern-title-root {
	min-width: 0;
	max-width: 100%;
}

.excluded-pattern-activator {
	min-width: 0;
	max-width: 100%;
	display: block;
}

.excluded-pattern-clamp {
	display: -webkit-box;
	-webkit-box-orient: vertical;
	line-clamp: 3;
	-webkit-line-clamp: 3;
	overflow: hidden;
	word-break: break-word;
	line-height: 1.4;
	max-height: calc(1.4em * 3);
}

.excluded-pattern-tooltip :deep(.tooltip-text) {
	white-space: pre-wrap;
}
</style>
