<template>
	<div class="mobile-title q-my-lg" v-if="platform === 'mobile'">
		{{ title }}
	</div>

	<q-bar class="bar" v-else>
		<q-icon v-if="icon" :name="icon" size="18px" />
		<div class="title col" :class="titAlign ? titAlign : 'text-left'">
			{{ title }}
		</div>
		<q-space />
		<q-btn
			dense
			flat
			icon="close"
			color="ink-3"
			v-close-popup
			@click="onCancel"
		>
			<q-tooltip>Close</q-tooltip>
		</q-btn>
	</q-bar>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';

defineProps({
	platform: {
		type: String,
		default: 'web',
		required: false
	},
	title: {
		type: String,
		default: 'Alert',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	titAlign: {
		type: String,
		default: 'left',
		required: false
	}
});

const $q = useQuasar();
</script>

<style scoped lang="scss">
.mobile-title {
	color: $ink-1;
	font-size: 16px;
	text-align: center;
	font-weight: 500;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}
.bar {
	background-color: $background-3;
	.title {
		height: 18px;
		line-height: 18px;
		color: $ink-1;
		font-size: 14px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}
</style>
