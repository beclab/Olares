<template>
	<q-menu
		@update:model-value="showPopupProxy"
		class="popup-menu bg-background-2"
	>
		<q-list dense padding>
			<template v-for="item in popupList" :key="item.action">
				<q-item
					class="row items-center justify-between text-ink-2 popup-item"
					clickable
					v-close-popup
					@click="handleEvent(item.action, item.type)"
				>
					<q-icon :name="item.icon" size="20px" class="q-mr-sm" />
					<q-item-section class="menuName"> {{ t(item.name) }}</q-item-section>

					<q-img
						v-if="item.active"
						style="width: 16px; height: 16px; margin-left: 20px"
						src="./../../../assets/images/active.svg"
					/>
				</q-item>
				<q-separator class="q-my-sm" v-if="item.action === 'mosaic'" />
			</template>
		</q-list>
	</q-menu>
</template>
<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDataStore } from '../../../stores/data';

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

const emits = defineEmits(['handleEvent']);

const store = useDataStore();

const { t } = useI18n();

const popupList = ref([
	{
		name: t('files.list_view'),
		icon: 'sym_r_dock_to_right',
		active: true,
		action: 'list',
		type: 'view'
	},
	{
		name: t('files.icon_view'),
		icon: 'sym_r_grid_view',
		active: false,
		action: 'mosaic',
		type: 'view'
	},

	{
		name: t('files.name'),
		icon: 'sym_r_sort_by_alpha',
		active: true,
		action: 'name',
		type: 'sort'
	},
	{
		name: t('files.file_type'),
		icon: 'sym_r_edit_calendar',
		active: false,
		action: 'type',
		type: 'sort'
	},
	{
		name: t('files.file_modified'),
		icon: 'sym_r_edit_document',
		active: false,
		action: 'file_modified',
		type: 'sort'
	},
	{
		name: t('files.file_size'),
		icon: 'sym_r_folder_copy',
		active: false,
		action: 'size',
		type: 'sort'
	}
]);

const showPopupProxy = (value: boolean) => {
	console.log(value);
};

const handleEvent = (vaule: string, type: string) => {
	emits('handleEvent', vaule);
	for (let i = 0; i < popupList.value.length; i++) {
		const popupItem = popupList.value[i];
		if (popupItem.type === type) {
			popupItem.active = false;
			if (popupItem.action === vaule) {
				popupItem.active = true;
			}
		}
	}
};

onMounted(() => {
	if (store.user.viewMode === 'list') {
		popupList.value[0].active = true;
		popupList.value[1].active = false;
	} else {
		popupList.value[1].active = true;
		popupList.value[0].active = false;
	}
});
</script>
<style lang="scss" scoped>
.popup-item {
	width: auto;
	border-radius: 4px;
}
.menuName {
	white-space: nowrap;
}
</style>
