<template>
	<div class="text-ink-2 text-subtitle1">{{ t('files.Accessible users') }}</div>
	<q-list class="full-width q-mt-md">
		<q-item v-for="item in users" :key="item.name" class="q-pa-none">
			<div class="row items-center justify-between" style="width: 100%">
				<div class="row items-center">
					<SMBUserIcon :name="item.name" />
					<div class="text-body1 text-ink-2 q-ml-sm">
						{{ item.name }}
					</div>
				</div>
				<div class="q-mr-md">
					<bt-select
						:offset="[30, 5]"
						v-model="item.editingPermission"
						:options="addDeleteOptions()"
						:height="40"
						classes="q-px-md"
						menuClasses="q-pa-xs"
						:menuItemHeight="40"
						@update:model-value="optionUpdate(item)"
						@update:menu="(status:boolean)=>{
							emits('updateMenu',status);
						}"
					>
					</bt-select>
				</div>
			</div>
		</q-item>
	</q-list>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import {
	SMBPermissionUser,
	smbPermissiontOpt
} from './../../../../stores/files';
import { SelectorProps } from 'src/constant';
import BtSelect from '../../../base/BtSelect.vue';
import SMBUserIcon from './SMBUserIcon.vue';

defineProps({
	users: {
		type: Array<SMBPermissionUser>,
		required: false,
		default: []
	}
});

const { t } = useI18n();

const emits = defineEmits(['delete', 'updateMenu']);

const addDeleteOptions = (): SelectorProps[] => {
	const options = smbPermissiontOpt() as SelectorProps[];
	options.push({
		label: t('delete'),
		value: 105,
		titleClass: 'text-negative'
	});

	return options;
};

const optionUpdate = (item: SMBPermissionUser) => {
	if ((item.editingPermission as any) == 105) {
		emits('delete', item);
	}
};
</script>

<style scoped lang="scss">
.owner {
	border-radius: 4px;
	height: 20px;
}
</style>
