<template>
	<div>
		<div class="text-h5 text-ink-1">{{ $t('bex.mnemonic_phrase') }}</div>
		<div class="q-mt-md">
			<AccountHeader></AccountHeader>
		</div>
		<div class="q-mt-sm">
			<DesktopMnemonics></DesktopMnemonics>
		</div>
	</div>
</template>

<script setup lang="ts">
import DesktopMnemonics from 'src/pages/Electron/SettingsPage/DesktopMnemonics.vue';
import AccountHeader from '../AccountHeader.vue';
</script>

<style scoped lang="scss"></style>
