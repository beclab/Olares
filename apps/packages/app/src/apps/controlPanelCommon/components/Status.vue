<template>
	<span class="row items-center">
		<MyBadge :type="type" />
		<span class="font-bold q-ml-xs">{{ name }}</span>
		<span v-if="!isNil(total) && !isNil(ready)">
			&nbsp;({{ ready }}/{{ total }})
		</span>
	</span>
</template>

<script setup lang="ts">
import { computed, defineProps } from 'vue';
import { isNaN, isNil } from 'lodash';
import MyBadge from './MyBadge.vue';
import _ from 'lodash';

interface StatusProps {
	name: string;
	total?: number;
	ready?: number;
	type: string;
	flicker?: boolean;
}

const props = withDefaults(defineProps<StatusProps>(), {});

const name = computed(() => _.capitalize(props.name) || '');

const total = isNaN(props.total) ? undefined : props.total;
const ready = isNaN(props.ready) ? undefined : props.ready;
</script>
