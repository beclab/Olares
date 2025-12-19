<template>
	<div class="row q-gutter-x-md">
		<q-btn
			outline
			color="light-blue-6 q-pa-none my-btn-wrapper"
			no-caps
			style="min-height: 24px"
		>
			<q-img :src="vaultPluginBtn1" :ratio="1" width="20px" spinner-size="xs" />
			<span class="q-ml-xs text-body3">Continue in Chat</span>
		</q-btn>
		<q-btn
			outline
			color="light-blue-6 q-pa-none my-btn-wrapper"
			no-caps
			style="min-height: 24px"
		>
			<q-img :src="vaultPluginBtn2" :ratio="1" width="20px" spinner-size="xs" />
			<span class="q-ml-xs text-body3">Add to Memo</span>
		</q-btn>
	</div>
</template>

<script setup lang="ts">
import vaultPluginBtn2 from 'assets/images/vault-plugin-btn-2.svg';
import vaultPluginBtn1 from 'assets/images/vault-plugin-btn-1.svg';
</script>

<style lang="scss" scoped>
.my-btn-wrapper {
	padding: 2px 8px;
	border-radius: 4px;
}
</style>
