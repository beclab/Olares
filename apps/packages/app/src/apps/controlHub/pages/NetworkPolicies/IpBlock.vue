<template>
	<div v-for="(item, index) in list" :key="index"></div>
</template>
<script setup lang="ts">
import { get, toPairs } from 'lodash-es';

const props = withDefaults(defineProps<{ data: { [key: string]: any } }>(), {});
const list = toPairs(props.data)
	.map((kv) => (kv[0] === 'cidr' ? `${kv[1]}` : kv.join(': ')))
	.reduce((acc: any, x: any) => (acc === null ? x : `${acc} ${x}`), null);
</script>

<style></style>
