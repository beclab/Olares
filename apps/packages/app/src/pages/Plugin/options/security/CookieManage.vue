<template>
	<div class="password-container">
		<div class="text-h5 text-ink-1">
			{{ $t('bex.cookies') }}
		</div>
		<CustomButton
			:label="$t('bex.manage_cookies')"
			color="yellow-default"
			class="q-px-lg q-mt-md"
			@click="clickHandler"
		>
		</CustomButton>
		<div class="column no-wrap flex-gap-y-sm q-mt-lg">
			<div class="text-body2 text-ink-2" style="width: 275px">
				{{ $t('bex.cookie_required_for') }}
			</div>
			<div class="text-body3 text-ink-3 row items-center no-wrap flex-gap-x-xs">
				<div class="dot"></div>
				{{ $t('bex.collecting_webpage') }}
			</div>
			<div class="text-body3 text-ink-3 row items-center no-wrap flex-gap-x-xs">
				<div class="dot"></div>
				{{ $t('bex.subscribe_personal_feed') }}
			</div>
			<div class="text-body3 text-ink-3 row items-center no-wrap flex-gap-x-xs">
				<div class="dot"></div>
				{{ $t('bex.downloading_resources') }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';
const userStore = useUserStore();

const clickHandler = () => {
	openUrl(
		userStore.getModuleSever('settings', 'https:', '/integration/cookie')
	);
};
</script>

<style lang="scss" scoped>
.password-container {
	.dot {
		position: relative;
		width: 16px;
		height: 16px;
		&::after {
			content: '';
			position: absolute;
			top: 50%;
			left: 50%;
			transform: translate(-50%, -50%);
			width: 8px;
			height: 8px;
			border-radius: 50%;
			background: #d9d9d9;
		}
	}
}
</style>
