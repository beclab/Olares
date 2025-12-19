<template>
	<q-icon
		class="cursor-pointer"
		:class="{ disabled: !appAbilitiesStore.wise.running }"
		name="sym_r_settings"
		color="ink-1"
		size="20px"
		@click="openWise"
	/>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';

const userStore = useUserStore();
const appAbilitiesStore = useAppAbilitiesStore();

const router = useRouter();

const openWise = () => {
	const url = userStore.getModuleSever(
		appAbilitiesStore.wise.id,
		'https:',
		`/edFeeds`
	);
	openUrl(url);
};
</script>

<style></style>
