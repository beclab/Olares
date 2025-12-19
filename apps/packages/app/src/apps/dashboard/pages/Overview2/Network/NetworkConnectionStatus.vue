<template>
	<div class="row items-center flex-gap-md">
		<q-img
			:src="dnsIcon"
			:ratio="1"
			spinner-size="0px"
			width="24px"
			height="24px"
		/>
		<template v-if="isWifi">
			<ConnectionProgress :connected="wifiConnected"></ConnectionProgress>
			<q-icon name="sym_r_wifi" size="24px" color="light-blue-default" />
			<ConnectionProgress
				:connected="wifiConnected && networkConnected"
			></ConnectionProgress>
		</template>
		<template v-else>
			<ConnectionProgress :connected="networkConnected"></ConnectionProgress>
		</template>
		<q-icon name="sym_r_language" size="24px" color="light-blue-default" />
	</div>
</template>

<script setup lang="ts">
import dnsIcon from '@apps/dashboard/assets/dns.svg';
import ConnectionProgress from './ConnectionProgress.vue';
interface Props {
	isWifi: boolean;
	networkConnected: boolean;
	wifiConnected?: boolean;
}

defineProps<Props>();
</script>
<style lang="scss" scoped>
.progress {
	height: 2px;
	flex: 1;
	background: linear-gradient(
		90deg,
		$light-blue-soft -0.85%,
		$light-blue-default 99.94%
	);
	&.progress-not {
		background: linear-gradient(90deg, $background-6 -0.85%, $ink-3 99.94%);
	}
}
</style>
