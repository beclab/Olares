<template>
	<q-layout view="hHh lpR fFf">
		<q-page-container>
			<router-view />
		</q-page-container>
	</q-layout>
</template>

<script lang="ts" setup>
import { useTokenStore } from '../stores/token';

const tokenStore = useTokenStore();
const host = window.location.origin;
tokenStore.setUrl(host);
</script>
