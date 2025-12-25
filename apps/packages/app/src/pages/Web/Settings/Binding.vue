<template>
	<div class="walletconnect_root">
		<div class="text-color-title text-subtitle1">
			{{ t(SETTING_MENU.rebinding.name) }}
		</div>

		<div class="row justify-center wrapper-content">
			<qrcode-vue :value="url" :size="400"> </qrcode-vue>

			<div class="text-color-sub-detail text-body3 bottom-content">
				{{ url }}
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import QrcodeVue from 'qrcode.vue';
import { SETTING_MENU } from '../../../utils/constants';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'BindingComponent',
	components: {
		QrcodeVue
	},
	setup() {
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

		const base_url =
			process.env.PL_SERVER_URL || window.location.origin + '/server';

		const bfl_token = ref<string>('');

		if (process.env.PL_SERVER_URL) {
			bfl_token.value = process.env.VAULT_MOCK_BFL_TOKEN;
		}

		const url = ref<string>(
			base_url +
				'?bfl_token=' +
				bfl_token.value +
				`&username=liuyu&password=${process.env.VAULT_MOCK_PASSWORD}`
		);

		const { t } = useI18n();

		return {
			url,
			options,
			tab,
			SETTING_MENU,
			t
		};
	}
});
</script>

<style lang="scss" scoped>
.walletconnect_root {
	width: 100%;

	.title {
		margin-top: 30px;
		text-align: center;
		width: 100%;
	}

	.wrapper-content {
		margin-top: 15px;
		height: 211px;
		width: 100%;
		word-wrap: break-word;
		word-break: break-all;
	}

	.bottom-content {
		margin-top: 15px;
		font-family: OpenSans;
		line-height: 18px;
		margin-left: 40px;
		margin-right: 40px;
	}
}
</style>
