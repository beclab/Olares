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
				<OrgInvitesList />
			</template>

			<template v-slot:after>
				<OrgInvitesView />
			</template>
		</q-splitter>
	</div>
</template>

<script lang="ts" setup>
import OrgInvitesView from './OrgInvitesView.vue';
import OrgInvitesList from './OrgInvitesList.vue';
import { useMenuStore } from 'src/stores/menu';

const meunStore = useMenuStore();

const updateSplitter = (value: number) => {
	meunStore.setSplitterModel(value);
};
</script>

<style lang="scss" scoped>
.itemWrap {
	width: 100%;
	height: 100%;
}

.borderRight {
	border-right: 1px solid $grey-12;
}
</style>
