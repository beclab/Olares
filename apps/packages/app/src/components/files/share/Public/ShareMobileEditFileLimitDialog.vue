<template>
	<bt-custom-dialog
		ref="CustomRef"
		:skip="false"
		:ok="false"
		:cancel="false"
		size="medium"
		position="bottom"
		platform="mobile"
		:content-pending="false"
	>
		<template v-for="item in diskUnitOptions()" :key="item.value">
			<terminus-item :show-board="false" @click="updateUnit(item.value)">
				<template v-slot:title>
					<div class="text-subtitle2 text-ink-2">
						{{ item.label }}
					</div>
				</template>
				<template v-slot:side v-if="item.value == unit">
					<div class="q-mr-lg">
						<q-icon name="sym_r_check" size="24px" color="ink-2" />
					</div>
				</template>
			</terminus-item>
		</template>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { PropType, ref } from 'vue';
import { DiskUnitMode, diskUnitOptions } from './public';
import TerminusItem from '../../../common/TerminusItem.vue';
const props = defineProps({
	unit: {
		type: Object as PropType<DiskUnitMode>,
		required: true
	}
});

const CustomRef = ref();

const onClose = () => {
	CustomRef.value.onDialogCancel();
};

// const permissionsOptions = [
// 	{
// 		name: sharePermissionStr(SharePermission.Edit),
// 		value: SharePermission.Edit,
// 		icon: 'sym_r_edit_note'
// 	},
// 	{
// 		name: sharePermissionStr(SharePermission.View),
// 		value: SharePermission.View,
// 		icon: 'sym_r_visibility'
// 	}
// ];

// const permissionClick = (permission: SharePermission, remove = false) => {
// 	if (permission == props.user.permission && !remove) {
// 		return;
// 	}
// 	CustomRef.value.onDialogOK({
// 		permission,
// 		remove
// 	});
// };

const updateUnit = (item: DiskUnitMode) => {
	if (item == props.unit) {
		return;
	}
	CustomRef.value.onDialogOK(item);
};
</script>

<style scoped lang="scss">
:deep(.dialog-content) {
	margin: 0px 0px 10px !important;
}
</style>
