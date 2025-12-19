<template>
	<div class="password-container">
		<div class="text-h5 text-ink-1">
			{{
				userStore.passwordReseted
					? t('change_local_password')
					: t(
							'This password is only used for unlocking {AppName} on this device',
							{
								AppName: 'LarePass'
							}
					  )
			}}
		</div>
		<CustomButton
			:label="
				userStore.passwordReseted
					? t('settings.changePassword')
					: t('Set up a password')
			"
			color="yellow-default"
			class="q-px-lg q-mt-md"
			@click="changePassword"
		>
		</CustomButton>
		<div class="column no-wrap flex-gap-y-sm q-mt-lg">
			<div class="text-body2 text-ink-2" style="width: 275px">
				{{ $t('bex.accounts_need_msg') }}
			</div>
			<div class="text-body3 text-ink-3 row items-center no-wrap flex-gap-x-xs">
				<div class="dot"></div>
				{{ $t('bex.accessing_vault') }}
			</div>
			<div class="text-body3 text-ink-3 row items-center no-wrap flex-gap-x-xs">
				<div class="dot"></div>
				{{ $t('bex.viewing_mnemonic') }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useTerminusSetSecurityPassword } from 'src/composables/mobile/useTerminusSetSecurityPassword';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
const { t, userStore, changePassword } = useTerminusSetSecurityPassword();
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
