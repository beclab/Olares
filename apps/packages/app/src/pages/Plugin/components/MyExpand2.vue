<template>
	<div>
		<div class="row items-center justify-between cursor-pointer q-pb-md">
			<span class="text-subtitle2 text-grey-10">{{ title }}</span>
			<div class="row items-center">
				<slot name="extra"></slot>
				<q-icon
					color="grey-5"
					size="16px"
					:name="isExpanded ? 'expand_less' : 'expand_more'"
					class="q-ml-xs"
					@click="isExpanded = !isExpanded"
				/>
			</div>
		</div>
		<div v-show="isExpanded" :duration="300">
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
interface Props {
	title?: string;
}
withDefaults(defineProps<Props>(), {});
const isExpanded = ref(true);
</script>
