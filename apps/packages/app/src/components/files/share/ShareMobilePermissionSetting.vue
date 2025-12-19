<template>
	<q-list class="permission-setting">
		<q-item v-for="item in users" :key="item.name" class="q-pa-none item">
			<div class="row items-center justify-between full-width">
				<div class="row items-center">
					<slot name="list-avatar" :user="item" />
					<div class="text-body1 text-ink-2 q-ml-sm">
						{{ item.name }}
					</div>
				</div>
				<div class="row items-center justify-end" @click="editPermission(item)">
					<div>{{ sharePermissionStr(item.permission) }}</div>
					<q-icon name="sym_r_expand_more" size="24px" color="ink-2" />
				</div>
			</div>
		</q-item>
	</q-list>
</template>

<script setup lang="ts">
import { sharePermissionStr } from 'src/stores/files';
import { SharePermission } from 'src/utils/interface/share';

interface SelectedUser {
	permission: SharePermission;
	name: string;
	olaresId?: string;
}

defineProps({
	users: {
		type: Array<SelectedUser>,
		required: false,
		default: []
	}
});

const editPermission = (user: SelectedUser) => {
	emit('editPermission', user);
};

const emit = defineEmits(['editPermission']);
</script>

<style scoped lang="scss">
.permission-setting {
	height: calc(75vh - 80px);
	.item {
		height: 56px;
	}
}
</style>
