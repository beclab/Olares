<template>
	<q-scroll-area
		ref="scrollAreaRef"
		:thumb-style="scrollBarStyle.thumbStyle"
		:class="keyboardOpen ? 'scroll-area-conf-open1' : 'scroll-area-conf-close1'"
	>
		<slot />
	</q-scroll-area>
</template>

<script setup lang="ts">
import { onMounted, watch } from 'vue';
import MonitorKeyboard from '../../utils/monitorKeyboard';
import { useQuasar } from 'quasar';
import { scrollBarStyle } from '../../utils/contact';
import { ref } from 'vue';
import { onUnmounted } from 'vue';

let monitorKeyboard: MonitorKeyboard | undefined = undefined;
const $q = useQuasar();
const keyboardOpen = ref(false);

onMounted(async () => {
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

const scrollAreaRef = ref();

watch(
	() => keyboardOpen.value,
	() => {
		if (keyboardOpen.value == true) {
			setTimeout(() => {
				scrollAreaRef.value.setScrollPosition('vertical', 1000, 300);
			}, 100);
		} else {
			scrollAreaRef.value.setScrollPosition('vertical', 0);
		}
	}
);
</script>

<style scoped lang="scss">
.scroll-area-conf-open1 {
	height: calc(100% - 120px);
	width: 100%;
	// padding-bottom: 10px;
}

.scroll-area-conf-close1 {
	height: 100%;
	width: 100%;
}
</style>
