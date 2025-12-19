<template>
	<div
		class="connect-layout row justify-center items-center q-ml-md"
		:class="!connected ? 'bg-red-1' : ''"
	>
		<div
			class="text-overline connect-title"
			:class="connected ? 'connected-title' : 'text-red-6'"
		>
			{{ label ? label : connected ? t('already_bound') : t('not_bound') }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
defineProps({
	connected: Boolean,
	label: String,
	icon: String
});

const { t } = useI18n();
</script>

<style scoped lang="scss">
.connect-layout {
	padding: 4px 8px;
	border-radius: 4px;
	width: auto;
	height: 20px;

	.connect-title {
		margin-left: 4px;
	}

	.connected-title {
		color: $blue;
	}
}
</style>
