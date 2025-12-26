<template>
	<div>
		<div @click="isExpanded = !isExpanded" class="cursor-pointer">
			<span class="text-body3 text-ink-2">{{
				isExpanded ? 'Click to collapse' : 'Click to expand'
			}}</span>
			<q-icon
				color="grey-5"
				size="16px"
				:name="isExpanded ? 'expand_less' : 'expand_more'"
				class="q-ml-xs"
			/>
		</div>
		<div v-show="isExpanded" :duration="300">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';

const isExpanded = ref(false);
</script>
