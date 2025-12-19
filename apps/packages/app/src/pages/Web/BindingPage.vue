<template>
	<div class="walletconnect_root">
		<div class="text-color-title text-subtitle1 title">Binding DID</div>
		<div class="row justify-center">
			<qrcode-vue :value="url" :size="400"> </qrcode-vue>
		</div>
		<div class="text-color-sub-detail text-body3 bottom-content">
			{{ url }}
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import QrcodeVue from 'qrcode.vue';
import { Encoder } from '@bytetrade/core';
import { useUserStore } from '../../stores/user';

export default defineComponent({
	name: 'BindingPage',
	components: {
		QrcodeVue
	},
	setup() {
		// const $q = useQuasar();
		const userStore = useUserStore();
		const options = ref([
			{
				value: 'QRCode',
				label: 'QR Code'
			},
			{
				value: 'URL',
				label: 'URL'
			}
		]);

		const tab = ref(options.value[0].value);

		const wizard = {
			url: process.env.PL_SERVER_URL || window.location.origin + '/server',
			username: userStore.terminusInfo()?.terminusName,
			password: '123456',
			system: {
				language: 'English',
				location: 'Singapore'
			},
			network: {
				enable_tunnel: false,
				external_ip: '127.0.0.1'
			}
		};
		const url = ref<string>(
			'active_vault://' + Encoder.stringToBase64Url(JSON.stringify(wizard))
		);

		return {
			url,
			options,
			tab
		};
	}
});
</script>

<style lang="scss" scoped>
.walletconnect_root {
	width: 100%;
	.title {
		text-align: center;
		width: 100%;
	}

	.bottom-content {
		width: 100%;
		line-height: 18px;
		word-break: break-all;
		word-wrap: break-word;
		text-align: left;
	}
}
</style>
