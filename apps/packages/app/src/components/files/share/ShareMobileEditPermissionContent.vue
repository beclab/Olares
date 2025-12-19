<template>
	<template v-for="item in permissions" :key="item.value">
		<terminus-item
			:show-board="false"
			:iconName="item.icon"
			:whole-picture-size="24"
			:icon-size="24"
			:item-height="48"
			@click="permissionClick(item.value)"
		>
			<template v-slot:title>
				<div class="text-subtitle2 text-ink-2">
					{{ item.name }}
				</div>
			</template>
			<template v-slot:side v-if="item.value == currentPermission">
				<div class="q-mr-lg">
					<q-icon name="sym_r_check" size="24px" color="ink-2" />
				</div>
			</template>
		</terminus-item>
	</template>
	<q-separator class="q-mx-lg" />

	<terminus-item
		:show-board="false"
		iconName="sym_r_delete"
		iconColor="negative"
		:whole-picture-size="24"
		:icon-size="24"
		:item-height="48"
		@click="permissionClick(currentPermission, true)"
	>
		<template v-slot:title>
			<div class="text-subtitle3-m text-negative">
				{{ t('delete') }}
			</div>
		</template>
	</terminus-item>
</template>

<script setup lang="ts">
import TerminusItem from '../../common/TerminusItem.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { sharePermissionStr } from '../../../stores/files';
import { SharePermission } from '../../../utils/interface/share';

const props = defineProps({
	currentPermission: {
		type: Number,
		required: true
	},
	permissions: {
		type: Array<{ name: string; value: SharePermission; icon: string }>,
		required: false,
		default: [
			{
				name: sharePermissionStr(SharePermission.ADMIN),
				value: SharePermission.ADMIN,
				icon: 'sym_r_manage_accounts'
			},
			{
				name: sharePermissionStr(SharePermission.Edit),
				value: SharePermission.Edit,
				icon: 'sym_r_edit_note'
			},
			{
				name: sharePermissionStr(SharePermission.View),
				value: SharePermission.View,
				icon: 'sym_r_visibility'
			}
		]
	}
});

const permissionClick = (permission: SharePermission, remove = false) => {
	if (permission == props.currentPermission && !remove) {
		return;
	}
	emits('update-permission', permission, remove);
};
const emits = defineEmits(['update-permission']);
</script>

<style scoped lang="scss"></style>
