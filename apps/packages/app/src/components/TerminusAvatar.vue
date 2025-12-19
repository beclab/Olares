<template>
	<q-avatar :size="`${size}px`">
		<img :src="imgRef" @error="imgOnError" />
		<q-badge floating rounded color="red" v-if="badge">{{ badge }}</q-badge>
	</q-avatar>
</template>
<script lang="ts" setup>
import { computed } from 'vue';

const props = defineProps({
	imageUri: {
		type: String,
		required: false
	},
	imagePath: {
		type: String,
		required: false
	},
	size: {
		type: Number,
		default: 32,
		required: false
	},
	badge: {
		type: String,
		default: '',
		required: false
	}
});

const imgRef = computed(() => {
	if (props.imageUri) {
		return props.imageUri;
	}
	if (props.imagePath) {
		return require(`../assets/${props.imagePath}`);
	}
	return require('../assets/account/default_avatar.svg');
});

function imgOnError(e) {
	const img = e.srcElement;
	img.src = require('../assets/account/default_avatar.svg');
	img.onerror = null;
}
</script>
