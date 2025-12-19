<template>
	<q-icon
		class="cursor-pointer"
		name="sym_r_settings"
		color="ink-1"
		size="20px"
		@click="linkToSetting(name, router)"
	/>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { RouteNames } from 'src/router/route-const';
import { linkToSetting } from 'src/utils/bex/link';
interface Props {
	name: RouteNames;
}

defineProps<Props>();

const router = useRouter();
</script>

<style></style>
