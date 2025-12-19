<template>
	<div
		class="progress"
		:class="[connected ? 'progress-yes' : 'progress-no']"
	></div>
	<q-icon
		:name="connected ? 'sym_r_check_circle' : 'sym_r_do_not_disturb_on'"
		size="24px"
		:color="connected ? 'light-blue-default' : 'ink-3'"
	/>
	<div
		class="progress"
		:class="[connected ? 'progress-yes' : 'progress-no']"
	></div>
</template>

<script setup lang="ts">
import dnsIcon from '@apps/dashboard/assets/dns.svg';
interface Props {
	connected?: boolean;
}
defineProps<Props>();
</script>
<style lang="scss" scoped>
.progress {
	height: 2px;
	flex: 1;
}
.progress-yes {
	background: linear-gradient(
		90deg,
		$light-blue-soft -0.85%,
		$light-blue-default 99.94%
	);
}
.progress-no {
	background: linear-gradient(90deg, $background-6 -0.85%, $ink-3 99.94%);
}
</style>
