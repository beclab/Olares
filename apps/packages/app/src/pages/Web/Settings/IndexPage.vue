<template>
	<q-splitter v-model="splitterModel" :limits="[30, 70]" style="height: 100vh">
		<template v-slot:before>
			<IndexList />
		</template>

		<template v-slot:after>
			<IndexView v-if="settingMode" :settingMode="settingMode" />
		</template>
	</q-splitter>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { useRoute } from 'vue-router';

import IndexView from './IndexView.vue';
import IndexList from './IndexList.vue';

const Route = useRoute();

const settingMode = ref('2');
const splitterModel = ref<number>(40);

watch(
	() => Route.params.mode,
	(newValue, oldValue) => {
		if (newValue == oldValue) {
			return;
		}
		settingMode.value = newValue as string;
	}
);
</script>

<style scoped lang="scss">
.setting {
	padding-top: 20px;

	.settingItem {
		height: 58px;
		line-height: 58px;
		border-bottom: 0.5px solid #ececec;
		box-sizing: border-box;

		&.itemActive {
			border-left: 2px solid $blue;
		}
	}
}
</style>
