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
			<ShareUserHeader :name="user?.name" @close="onClose">
				<template v-slot:avatar>
					<q-avatar :size="`32px`" class="">
						<TerminusAvatar
							:info="{
								terminusName: user?.olaresId
							}"
							:size="32"
						/>
					</q-avatar>
				</template>
			</ShareUserHeader>
			<q-separator />
		</template>
		<ShareMobileEditPermissionContent
			:current-permission="user.permission"
			@update-permission="permissionClick"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import ShareUserHeader from '../ShareUserHeader.vue';
import { PropType, ref } from 'vue';
import { SharePermission } from '../../../../utils/interface/share';

import ShareMobileEditPermissionContent from '../ShareMobileEditPermissionContent.vue';

interface SelectedUser {
	permission: SharePermission;
	name: string;
	olaresId: string;
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
