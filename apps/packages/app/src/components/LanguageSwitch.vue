<template>
	<q-btn color="primary" icon="check" label="OK" @click="onClick" />
</template>

<script setup lang="ts">
import { i18n } from 'src/boot/control-hub-i18n';
const onClick = () => {
	const en = 'en-US';
	const zh = 'zh-CN';
	i18n.global.locale.value = i18n.global.locale.value === en ? zh : en;
};
</script>

<style></style>
