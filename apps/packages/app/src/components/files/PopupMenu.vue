<template>
	<q-menu
		@update:model-value="showPopupProxy"
		class="popup-menu bg-background-2"
	>
		<q-list dense padding>
			<q-item
				class="row items-center justify-between text-ink-3 popup-item"
				dense
			>
				{{ t('main.sort_by') }}
			</q-item>
			<template v-for="item in popupList" :key="item.action">
				<q-item
					class="row items-center justify-between text-ink-2 popup-item"
					clickable
					v-close-popup
					@click="handleEvent(item)"
				>
					<q-icon :name="item.icon" size="20px" class="q-mr-sm" />
					<q-item-section class="menuName">
						{{ t(`files.file_${item.name}`) }}</q-item-section
					>
					<q-img
						v-if="item.active"
						class="q-mr-sm"
						style="width: 16px; height: 16px"
						src="./../../assets/images/active.svg"
					/>
				</q-item>
			</template>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
defineProps({
	item: {
		type: Object,
		required: false
	},
	from: {
		type: String,
		require: false,
		default: ''
	},
	isSide: {
		type: Boolean,
		require: false,
		default: false
	}
});

const emits = defineEmits(['handleEvent', 'popupState']);

const { t } = useI18n();

const popupList = ref([
	{
		name: 'name',
		icon: 'sym_r_grid_view',
		active: true,
		action: 'name',
		type: 'sort'
	},
	{
		name: 'type',
		icon: 'sym_r_edit_calendar',
		active: false,
		action: 'type',
		type: 'sort'
	},
	{
		name: 'modified',
		icon: 'sym_r_edit_document',
		active: false,
		action: 'modified',
		type: 'sort'
	},
	{
		name: 'size',
		icon: 'sym_r_folder_copy',
		active: false,
		action: 'size',
		type: 'sort'
	}
]);

const showPopupProxy = (value: boolean) => {
	emits('popupState', value);
};

const handleEvent = (item) => {
	for (let i = 0; i < popupList.value.length; i++) {
		const popupItem = popupList.value[i];
		if (popupItem.type === item.type) {
			popupItem.active = false;
			if (popupItem.action === item.action) {
				popupItem.active = true;
				emits('handleEvent', item);
			}
		}
	}
};
</script>
<style lang="scss" scoped>
.popup-item {
	width: 150px;
	border-radius: 4px;
	padding: 8px 4px;
}
.menuName {
	white-space: nowrap;
}
</style>
