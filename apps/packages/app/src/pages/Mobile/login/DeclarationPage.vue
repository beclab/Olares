<template>
	<div class="column justify-center items-center full-height">
		<q-img src="/larepass/LarePass.png" class="termipass-logo" />
		<div class="termipass-tips text-subtitle1 text-ink-1 q-mt-lg">
			{{ t('VPN used to secure connections to Olares') }}
		</div>
		<q-btn
			class="bg-yellow-default text-ink-on-brand-black-light text-subtitle1 confirm-btn"
			:label="t('Get Started')"
			@click="router.replace('/welcome')"
		/>
		<div class="text-body3 text-ink-2 termipass-desc">
			{{
				t(
					'LarePass uses VPN services to ensure user privacy and security. The VPN connects your device to the network of the Olares server, allowing access to its resources. The service will not access, hijack, or tamper with your content.'
				)
			}}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

const router = useRouter();
const { t } = useI18n();
</script>

<style scoped lang="scss">
.termipass-logo {
	width: 72px;
	height: 72px;
	border-radius: 50%;
}

.termipass-tips {
	margin-left: 45px;
	margin-right: 45px;
	width: calc(100% - 90px);
	text-align: center;
}

.confirm-btn {
	width: 207px;
	height: 48px;
	border-radius: 8px;
	margin-top: 64px;
	text-transform: capitalize;
}

.termipass-desc {
	margin-top: 64px;
	margin-left: 32px;
	margin-right: 32px;
	width: calc(100% - 64px);
	text-align: center;
}
</style>
