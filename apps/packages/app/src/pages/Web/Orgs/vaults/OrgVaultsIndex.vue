<template>
	<div class="row itemWrap items-start justify-between">
		<q-splitter
			v-model="meunStore.splitterModel"
			style="height: 100%; width: 100%"
			:limits="[40, 60]"
			separator-class="separator-class"
			before-class="before-class"
			after-class="after-class"
			@update:model-value="updateSplitter"
		>
			<template v-slot:before>
				<OrgVaultsList />
			</template>

			<template v-slot:after>
				<OrgVaultsView />
			</template>
		</q-splitter>
	</div>
</template>

<script lang="ts" setup>
import OrgVaultsList from './OrgVaultsList.vue';
import OrgVaultsView from './OrgVaultsView.vue';
import { useMenuStore } from 'src/stores/menu';
const meunStore = useMenuStore();

const updateSplitter = (value: number) => {
	meunStore.setSplitterModel(value);
};
</script>

<style lang="scss" scoped>
.itemWrap {
	height: 100%;
}
</style>
