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
		<template v-slot:header>
			<ShareUserHeader :name="user.name" @close="onClose">
				<template v-slot:avatar>
					<SMBUserIcon :name="user.name" />
				</template>
			</ShareUserHeader>
			<q-separator />
		</template>
		<ShareMobileEditPermissionContent
			:permissions="permissionsOptions"
			:current-permission="user.permission"
			@update-permission="permissionClick"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import ShareUserHeader from '../ShareUserHeader.vue';
import { PropType, ref } from 'vue';
import { SharePermission } from '../../../../utils/interface/share';
import SMBUserIcon from './SMBUserIcon.vue';
import ShareMobileEditPermissionContent from '../ShareMobileEditPermissionContent.vue';
import { sharePermissionStr } from 'src/stores/files';

interface SelectedUser {
	permission: SharePermission;
	name: string;
}

const props = defineProps({
	user: {
		type: Object as PropType<SelectedUser>,
		required: true
	}
});

const CustomRef = ref();

const onClose = () => {
	CustomRef.value.onDialogCancel();
};

const permissionsOptions = [
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
];

const permissionClick = (permission: SharePermission, remove = false) => {
	if (permission == props.user.permission && !remove) {
		return;
	}
	CustomRef.value.onDialogOK({
		permission,
		remove
	});
};
</script>

<style scoped lang="scss">
:deep(.dialog-content) {
	margin: 0px 0px 10px !important;
}
</style>
