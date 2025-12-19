<template>
	<div class="column no-wrap flex-gap-y-sm full-width">
		<div class="row items-center flex-gap-x-sm">
			<span class="text-subtitle1 text-ink-1">{{ title }}</span>
			<div class="row items-center relative-position">
				<q-icon v-if="info" name="sym_r_info" color="ink-3" size="20px" />
				<BtTooltip
					:label="info"
					anchor="center right"
					self="center left"
					:offset="[5, 0]"
					style="width: 240px; white-space: wrap; text-align: left"
				/>
			</div>
		</div>
		<div>
			<slot></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import BtTooltip from '../base/BtTooltip.vue';

interface Props {
	title: string;
	info?: string;
}

defineProps<Props>();
</script>

<style></style>
