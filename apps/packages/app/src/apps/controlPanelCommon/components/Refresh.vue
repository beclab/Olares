<template>
	<QButtonStyle>
		<q-btn dense flat icon="sym_r_refresh" @click="clickHandler" />
	</QButtonStyle>
</template>

<script setup lang="ts">
import QButtonStyle from './QButtonStyle.vue';

const emits = defineEmits<{
	(event: 'click'): void;
}>();
const clickHandler = () => {
	emits('click');
};
</script>
