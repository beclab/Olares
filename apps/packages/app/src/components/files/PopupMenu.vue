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
						v-if="item.type == activedSort"
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
import { FilesSortType } from './../../utils/contact';
import { filesSortOptions } from 'src/utils/interface/files';

withDefaults(
	defineProps<{
		activedSort?: FilesSortType;
	}>(),
	{
		activedSort: FilesSortType.Modified
	}
);

const emits = defineEmits(['handleEvent', 'popupState']);

const { t } = useI18n();

const popupList = ref(filesSortOptions);

const showPopupProxy = (value: boolean) => {
	emits('popupState', value);
};

const handleEvent = (item) => {
	emits('handleEvent', item);
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
