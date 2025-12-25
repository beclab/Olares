<template>
	<div class="text-ink-2 text-subtitle1">{{ t('files.Accessible users') }}</div>
	<q-list class="full-width q-mt-md">
		<q-item v-for="item in sortUser" :key="item.name" class="q-pa-none">
			<div class="row items-center justify-between" style="width: 100%">
				<div class="row items-center">
					<q-avatar :size="`32px`" class="">
						<TerminusAvatar
							:info="{
								terminusName: item.olaresId
							}"
							:size="32"
						/>
					</q-avatar>
					<div class="text-body1 text-ink-2 q-ml-sm">
						{{ item.name }}
					</div>
					<div
						v-if="item.isOwner"
						class="owner bg-background-3 text-light-blue-default text-subtitle3 q-ml-md row items-center q-px-sm"
					>
						{{ t('owner') }}
					</div>
				</div>
				<div v-if="item.isOwner" class="q-mr-md">
					{{ t('admin') }}
				</div>
				<div
					v-else-if="item.permission == SharePermission.ADMIN && !isOwner"
					class="q-mr-md"
				>
					{{ sharePermissionStr(item.permission) }}
				</div>
				<div
					v-else-if="
						users.find(
							(e) =>
								e.name == currentUser && e.permission == SharePermission.ADMIN
						)
					"
					class="q-mr-md"
				>
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
					/>
				</div>
				<div v-else class="q-mr-md">
					{{ sharePermissionStr(item.permission) }}
				</div>
			</div>
		</q-item>
	</q-list>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';

import {
	ShareItemUser,
	permissionOpt,
	sharePermissionStr
} from './../../../../stores/files';
import { SharePermission } from 'src/utils/interface/share';
import { OLARES_ROLE, SelectorProps } from 'src/constant';
import BtSelect from '../../../base/BtSelect.vue';
import { computed } from 'vue';

const props = defineProps({
	users: {
		type: Array<ShareItemUser>,
		required: false,
		default: []
	},
	currentUser: {
		type: String,
		required: true
	}
});

const { t } = useI18n();

const emits = defineEmits(['delete', 'updateMenu']);

const addDeleteOptions = (): SelectorProps[] => {
	let options = permissionOpt() as SelectorProps[];

	if (!isOwner.value) {
		options = options.filter((e) => e.value != SharePermission.ADMIN);
	}

	options.push({
		label: t('delete'),
		value: 105,
		titleClass: 'text-negative'
	});

	return options;
};

const optionUpdate = (item: ShareItemUser) => {
	if ((item.editingPermission as any) == 105) {
		emits('delete', item);
	}
};

const isOwner = computed(() => {
	return props.users.find((e) => e.name == props.currentUser && e.isOwner);
});

const sortUser = computed(() => {
	return props.users
		.filter((e) => e.isOwner)
		.concat(props.users.filter((e) => !e.isOwner));
});
</script>

<style scoped lang="scss">
.owner {
	border-radius: 4px;
	height: 20px;
}
</style>
