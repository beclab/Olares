<template>
	<div>
		<component
			ref="currentComponent"
			:origin_id="origin_id"
			:is="currentComponent"
		/>
	</div>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from 'vue';
import { useDataStore } from '../../../stores/data';

import Help from './Help.vue';
import Info from './InfoDialog.vue';
import Delete from './DeleteDialog.vue';
import Rename from './RenameDialog.vue';
import NewDir from './NewDir.vue';
import NewLib from './NewLib.vue';
import InternalShare from '../share/Internal/InternalShare.vue';
import InternalMobileShare from '../share/Internal/InternalMobileShare.vue';
import SMBShare from '../share/SMB/SMBShare.vue';
import PublicShare from '../share/Public/PublicShare.vue';
import SMBMobileShare from '../share/SMB/SMBMobileShare.vue';
import PublicMobileShare from '../share/Public/PublicMobileShare.vue';
import ShareResetPassword from '../share/ShareResetPassword.vue';
import RevokeShare from '../share/RevokeShareDialog.vue';

defineProps({
	origin_id: {
		type: Number,
		required: true
	}
});

const store = useDataStore();
const show = ref<null | string>(null);

let actionList = [
	'newDir',
	'NewLib',
	'info',
	'rename',
	'delete',
	'help',
	'share-internal-dialog',
	'share-internal-mobile-dialog',
	'share-smb-dialog',
	'share-smb-mobile-dialog',
	'share-public-dialog',
	'share-public-mobile-dialog',
	'share-reset-password',
	'revoke_share'
];
let componentList = [
	NewDir,
	NewLib,
	Info,
	Rename,
	Delete,
	Help,
	InternalShare,
	InternalMobileShare,
	SMBShare,
	SMBMobileShare,
	PublicShare,
	PublicMobileShare,
	ShareResetPassword,
	RevokeShare
];
watch(
	() => store.show,
	(newVal, oldVal) => {
		if (oldVal == newVal) {
			return;
		}
		show.value = newVal;
	}
);

const currentComponent = computed(function () {
	const matched = actionList.indexOf(show.value || '');

	if (matched >= 0 && show.value) {
		return componentList[matched];
	}
	return null;
});
</script>
