<template>
	<div class="web3-common-header">
		<div class="notification-title-class row justify-center items-center">
			<div class="text-ink-1 text-subtitle1">{{ title }}</div>
		</div>
		<div v-show="showDappInfo" class="dapp-info fit column items-center">
			<q-img :src="getRequireImage(dappLogo)" height="44px" width="44px" />
			<div class="text-color-title text-subtitle2 dapp-url">{{ dappUrl }}</div>
			<div class="text-ink-2 text-subtitle1 dapp-connect-title">
				{{ action }}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { getRequireImage } from '../../../utils/imageUtils';

withDefaults(
	defineProps<{
		title: string;
		dappLogo: string;
		dappUrl: string;
		action: string;
		showDappInfo: boolean;
	}>(),
	{
		title: '',
		dappLogo: '',
		dappUrl: '',
		action: '',
		showDappInfo: false
	}
);
</script>

<style lang="scss" scoped>
.web3-common-header {
	width: 100%;

	.notification-title-class {
		height: 74px;
		width: 100%;
	}

	.dapp-info {
		margin-bottom: 16px;

		.dapp-url {
			margin-top: 12px;
		}

		.dapp-connect-title {
			margin-top: 16px;
		}
	}
}
</style>
