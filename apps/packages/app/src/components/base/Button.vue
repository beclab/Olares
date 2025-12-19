<template>
	<button class="bt-button">
		<slot></slot>
	</button>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
	name: 'BtButton'
});
</script>
