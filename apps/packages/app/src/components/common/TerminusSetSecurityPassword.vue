<template>
	<div>
		<div class="text-h6 text-ink-1">
			{{ t('settings.safety') }}
		</div>
		<div class="text-ink-2 text-body2 q-mt-md">
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
		<div class="adminBtn q-mt-md text-body3" @click="changePassword">
			<q-icon name="sym_r_lock" class="icon q-mr-xs" size="16px" />
			{{
				userStore.passwordReseted
					? t('settings.changePassword')
					: t('Set up a password')
			}}
		</div>
	</div>
</template>

<script setup lang="ts">
import { useTerminusSetSecurityPassword } from 'src/composables/mobile/useTerminusSetSecurityPassword';

const { t, userStore, changePassword } = useTerminusSetSecurityPassword();
</script>

<style lang="scss" scoped>
.adminBtn {
	border: 1px solid $yellow;
	background-color: $yellow;
	display: inline-block;
	color: $grey-10;
	padding: 7px 12px;
	border-radius: 8px;
	cursor: pointer;

	&:hover {
		background-color: $yellow-3;
	}
}
</style>
