<template>
	<div class="continer">
		<q-splitter
			v-model="meunStore.splitterModel"
			style="height: 100%"
			:limits="[40, 60]"
			separator-class="separator-class"
			before-class="before-class"
			after-class="after-class"
			@update:model-value="updateSplitter"
		>
			<template v-slot:before>
				<ItemList @toolabClick="onItemClicked" />
			</template>

			<template v-slot:after>
				<ItemView
					ref="itemViewRef"
					v-if="itemID"
					:itemID="itemID"
					:isNew="meunStore.isEdit"
				/>

				<div
					class="text-ink-2 column items-center justify-center full-height noData bg-background-1"
					v-if="!itemID"
				>
					<img class="q-mt-sm" src="../../assets/layout/nodata.svg" />
					<span class="q-mb-md text-gray-5" style="margin-top: 32px">
						{{ t('vault_t.no_item_selected') }}
					</span>
				</div>
			</template>
		</q-splitter>
	</div>
</template>

<script lang="ts" setup>
// import { useQuasar } from 'quasar';
import { watch, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useMenuStore } from '../../stores/menu';
import ItemView from './ItemView.vue';
import ItemList from './ItemList.vue';
import { useI18n } from 'vue-i18n';

const itemID = ref();
const Router = useRouter();
const Route = useRoute();
const meunStore = useMenuStore();
const itemViewRef = ref();
const { t } = useI18n();

function onItemClicked(itemid: string) {
	if (itemViewRef.value?.editing_t1) {
		const answer = window.confirm(
			t(
				'vault_t.are_you_sure_you_want_to_leave_this_page_any_changes_will_be_lost'
			)
		);
		if (answer) {
			Router.push({
				path: '/items/' + (itemid ? itemid : '')
			});
		}
	} else {
		Router.push({
			path: '/items/' + (itemid ? itemid : '')
		});
	}
}

const updateSplitter = (value: number) => {
	meunStore.setSplitterModel(value);
};

watch(
	() => Route.params.itemid,
	(newValue, oldValue) => {
		if (newValue == oldValue) {
			return;
		}
		if (!newValue) {
			itemID.value = null;
			return false;
		}

		itemID.value = newValue;
	}
);
</script>
<style lang="scss">
.separator-class {
	width: 1px;
	background: transparent !important;
}
</style>

<style lang="scss" scoped>
.continer {
	width: 100%;
	height: 100%;
	position: absolute;
	left: 0;
	top: 0;
}

.noData {
	width: 100%;
	height: 100%;
	display: flex;
	flex-direction: column;
	overflow: hidden;
	float: right;
}
</style>
