<template>
	<q-card class="user-info">
		<div class="header q-pa-md row items-center justify-between">
			<div class="row items-center title text-ink=1">
				<q-icon name="sym_r_tune" size="20px" />
				<div class="text-subtitle2 q-ml-xs">{{ t('files.Filter') }}</div>
			</div>
		</div>

		<q-list class="q-px-md content">
			<mutiple-select :options="shared" :title="t('files.Shared')" />
			<mutiple-select :options="owner" :title="t('owner')" class="q-mt-md" />

			<mutiple-select
				:options="scope"
				:title="t('files.Share scope')"
				class="q-mt-md"
			/>

			<mutiple-select
				:options="permission"
				:title="t('files.permission')"
				class="q-mt-md"
			/>

			<q-item class="q-pa-none item q-mt-md">
				<single-select
					:title="t('files.Expiration date')"
					:offset="[0, 0]"
					v-model="expirTime"
					:options="expirTimeSelect"
					:height="40"
					classes="q-px-md"
					menuClasses="q-pa-xs"
					:menuItemHeight="40"
				></single-select>
			</q-item>
		</q-list>
		<q-card-actions class="row items-center justify-between footer q-px-md">
			<q-item
				clickable
				dense
				class="but-cancel text-body3 ink-2 row justify-center items-center"
				@click="resetall"
			>
				{{ t('files.Reset all') }}
			</q-item>
			<q-item
				clickable
				dense
				:class="'but-creat'"
				class="text-body3 yellow-default ink-on-brand-black row justify-center items-center"
				@click="confirmFilter"
			>
				{{ t('files.Apply') }}
			</q-item>
		</q-card-actions>
	</q-card>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import {
	useFilesStore,
	ExoirationTime,
	shareTypeStr,
	sharePermissionStr
} from 'src/stores/files';
import { ShareType, SharePermission } from 'src/utils/interface/share';
import MutipleSelect from '../filter/MutipleSelect.vue';
import SingleSelect from '../filter/SingleSelect.vue';

import { ref } from 'vue';
import { DriveType } from 'src/utils/interface/files';
import { useRoute } from 'vue-router';

const { t } = useI18n();

const emit = defineEmits(['close']);

const filesStore = useFilesStore();
const route = useRoute();

const expirTime = ref(filesStore.shareFilter.expire);

const all = {
	value: '',
	label: t('my.all'),
	selected: true,
	isAll: true,
	isDefault: true
};

const shared = ref([
	{
		...all,
		selected:
			filesStore.shareFilter.shared.byMe && filesStore.shareFilter.shared.withMe
	},
	{
		value: 1,
		label: t('files.Shared By me'),
		selected: filesStore.shareFilter.shared.byMe,
		isAll: false,
		isDefault: false
	},
	{
		value: 2,
		label: t('files.Shared With me'),
		selected: filesStore.shareFilter.shared.withMe,
		isAll: false,
		isDefault: false
	}
]);

const owner = ref([
	{ ...all, selected: filesStore.shareFilter.owner.length > 0 }
]);

if (filesStore.users) {
	owner.value = owner.value.concat(
		filesStore.users.users.map((e) => {
			return {
				value: e.name,
				label: e.name,
				selected: filesStore.shareFilter.owner.includes(e.name),
				isAll: false,
				isDefault: false
			};
		})
	);
}

const scope = ref([
	{
		...all,
		selected:
			filesStore.shareFilter.scope.public &&
			filesStore.shareFilter.scope.smb &&
			filesStore.shareFilter.scope.internal
	},
	{
		value: ShareType.PUBLIC,
		label: shareTypeStr(ShareType.PUBLIC),
		selected: filesStore.shareFilter.scope.public,
		isAll: false,
		isDefault: false
	},
	{
		value: ShareType.SMB,
		label: shareTypeStr(ShareType.SMB),
		selected: filesStore.shareFilter.scope.smb,
		isAll: false,
		isDefault: false
	},
	{
		value: ShareType.INTERNAL,
		label: shareTypeStr(ShareType.INTERNAL),
		selected: filesStore.shareFilter.scope.internal,
		isAll: false,
		isDefault: false
	}
]);

const permission = ref([
	{
		...all,
		selected:
			filesStore.shareFilter.permission.manage &&
			filesStore.shareFilter.permission.edit &&
			filesStore.shareFilter.permission.view
	},
	{
		value: SharePermission.ADMIN,
		label: sharePermissionStr(SharePermission.ADMIN),
		selected: filesStore.shareFilter.permission.manage,
		isAll: false,
		isDefault: false
	},
	{
		value: SharePermission.Edit,
		label: sharePermissionStr(SharePermission.Edit),
		selected: filesStore.shareFilter.permission.edit,
		isAll: false,
		isDefault: false
	},
	{
		value: SharePermission.View,
		label: sharePermissionStr(SharePermission.View),
		selected: filesStore.shareFilter.permission.view,
		isAll: false,
		isDefault: false
	}
]);

const expirTimeSelect = ref([
	{
		value: ExoirationTime.all,
		label: t('my.all')
	},
	{
		value: ExoirationTime.within1days,
		label: t('files.With in 1 day')
	},
	{
		value: ExoirationTime.within7days,
		label: t('files.With in {number} days', {
			number: 7
		})
	},
	{
		value: ExoirationTime.within30days,
		label: t('files.With in {number} days', {
			number: 30
		})
	},
	{
		value: ExoirationTime.within1year,
		label: t('files.With in 1 year')
	},
	{
		value: ExoirationTime.over1year,
		label: t('files.Over 1 year')
	}
]);

const resetall = () => {
	filesStore.resetShareFilter();
	filesStore.setBrowserUrl(route.fullPath, DriveType.Share);
	emit('close');
};

const confirmFilter = () => {
	filesStore.shareFilter.shared = {
		byMe: shared.value[1].selected,
		withMe: shared.value[2].selected
	};

	filesStore.shareFilter.owner = owner.value
		.filter((e) => e.selected && !e.isAll)
		.map((e) => e.label);

	filesStore.shareFilter.scope = {
		public: scope.value[1].selected,
		smb: scope.value[2].selected,
		internal: scope.value[3].selected
	};

	filesStore.shareFilter.permission = {
		manage: permission.value[1].selected,
		edit: permission.value[2].selected,
		view: permission.value[3].selected
	};

	filesStore.shareFilter.expire = expirTime.value;

	filesStore.setBrowserUrl(route.fullPath, DriveType.Share);
	emit('close');
};
</script>

<style scoped lang="scss">
.user-info {
	width: 320px;
	height: 460px;
	border-radius: 8px;

	.header {
		height: 44px;
		width: 100%;

		.title {
			flex: 1;
		}
	}

	.content {
		height: 360px;
		.item {
			height: 60px;
		}
	}
	.footer {
		width: 100%;
		height: 56px;
		.but-creat {
			border-radius: 8px;
			background: $yellow;
			color: $grey-10;
			width: 48%;
		}

		.but-cancel {
			border-radius: 8px;
			width: 48%;
			border: 1px solid $btn-stroke;
		}
	}
}
</style>
