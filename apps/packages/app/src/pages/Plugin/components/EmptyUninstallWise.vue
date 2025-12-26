<template>
	<EmptyData
		:title="$t('bex.unable_access_wise')"
		class="absolute-center"
		@click="appAbilitiesStore.refreshAppAbilities"
	>
		<template #subtitle>
			<span class="text-blue-default" @click="() => openWiseInMarket()">{{
				$t('bex.install_wise_in_collect_page')
			}}</span>
		</template>
	</EmptyData>
	<q-inner-loading :showing="appAbilitiesStore.loading"> </q-inner-loading>
</template>

<script setup lang="ts">
import EmptyData from './EmptyData.vue';
import { useWiseAbility } from 'src/composables/common/useWiseAbility';
const { openWiseInMarket, appAbilitiesStore } = useWiseAbility();
</script>

<style></style>
