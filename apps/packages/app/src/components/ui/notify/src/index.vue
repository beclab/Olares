<template>
	<q-scroll-area
		ref="scrollRef"
		v-bind="$attrs"
		:thumb-style="
			$attrs['thumb-style'] ? $attrs['thumb-style'] : scrollBarStyle.thumbStyle
		"
	>
		<slot />
	</q-scroll-area>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const scrollRef = ref();

const scrollBarStyle = ref({
	contentStyle: {},
	contentActiveStyle: {},
	thumbStyle: {
		right: '2px',
		borderRadius: '3px',
		backgroundColor: '#BCBDBE',
		width: '6px',
		opacity: '1'
	}
});
</script>

<script lang="ts">
import { defineComponent } from 'vue';
export default defineComponent({
	name: 'BtScrollArea'
});
</script>
