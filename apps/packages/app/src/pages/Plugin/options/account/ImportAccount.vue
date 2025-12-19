<template>
	<div class="input-mnemonic-bex-wrapper">
		<InputMnemonicPage></InputMnemonicPage>
	</div>
</template>

<script setup lang="ts">
import InputMnemonicPage from 'src/pages/Electron/import/InputMnemonicPage.vue';
</script>
<style lang="scss" scoped>
.input-mnemonic-bex-wrapper {
	height: 590px;
	border: 1px solid $separator-2;
	border-radius: 12px;
	overflow: hidden;
}
</style>
