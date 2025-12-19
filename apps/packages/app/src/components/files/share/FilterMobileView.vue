<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.Filter')"
		:skip="false"
		:ok="false"
		:cancel="false"
		platform="mobile"
		size="medium"
		position="bottom"
	>
		<mobile-select-header :title="t('files.Shared')" />

		<MobileBaseSelect :options="shared" :row-count="3" />

		<mobile-select-header
			:title="t('files.Owner')"
			class="q-mt-sm"
			:showExtend="owner.values.length > 8"
			:expand="owner.expand"
			@revertExpand="owner.expand = !owner.expand"
		/>

		<MobileBaseSelect
			:options="
				owner.values.length > 8 && !owner.expand
					? owner.values.slice(0, 8)
					: owner.values
			"
			:row-count="4"
			:singleSelect="false"
		/>

		<mobile-select-header :title="t('files.Share scope')" class="q-mt-sm" />

		<MobileBaseSelect :options="scope" :row-count="4" />

		<mobile-select-header :title="t('files.permission')" class="q-mt-sm" />

		<MobileBaseSelect :options="permission" :row-count="4" />

		<mobile-select-header
			:title="t('files.Expiration date')"
			class="q-mt-sm"
			:showExtend="true"
			:expand="expirTimeSelect.expand"
			@revertExpand="expirTimeSelect.expand = !expirTimeSelect.expand"
		/>

		<MobileBaseSelect
			:options="
				!expirTimeSelect.expand
					? expirTimeSelect.values.slice(0, 5)
					: expirTimeSelect.values
			"
		/>
		<q-card-actions class="row items-center justify-between footer q-px-md">
			<q-item
				clickable
				dense
				class="but-cancel btn-item text-body3 ink-2 row justify-center items-center"
				@click="resetall"
			>
				{{ t('files.Reset all') }}
			</q-item>
			<q-item
				clickable
				dense
				:class="'but-creat'"
				class="btn-item text-body3 yellow-default ink-on-brand-black row justify-center items-center"
				@click="confirmFilter"
			>
				{{ t('confirm') }}
			</q-item>
		</q-card-actions>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import MobileSelectHeader from '../filter/MobileSelectHeader.vue';
import MobileBaseSelect from '../filter/MobileBaseSelect.vue';
import { ref } from 'vue';
import {
	ExoirationTime,
	sharePermissionStr,
	shareTypeStr,
	useFilesStore
} from 'src/stores/files';
import { SharePermission, ShareType } from 'src/utils/interface/share';
import { DriveType } from 'src/utils/interface/files';
import { useRoute } from 'vue-router';

const { t } = useI18n();

const filesStore = useFilesStore();

const route = useRoute();

const CustomRef = ref();

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
		selected:
			filesStore.shareFilter.shared.byMe &&
			!filesStore.shareFilter.shared.withMe,
		isAll: false,
		isDefault: false
	},
	{
		value: 2,
		label: t('files.Shared With me'),
		selected:
			filesStore.shareFilter.shared.withMe &&
			!filesStore.shareFilter.shared.byMe,
		isAll: false,
		isDefault: false
	}
]);

const owner = ref({
	expand: false,
	values: [{ ...all, selected: filesStore.shareFilter.owner.length > 0 }]
});

if (filesStore.users) {
	const mockUser = [
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles',
		// 'Charles'
	];
	owner.value.values = owner.value.values.concat(
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
	owner.value.values = owner.value.values.concat(
		mockUser.map((e) => {
			return {
				value: e,
				label: e,
				selected: filesStore.shareFilter.owner.includes(e),
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
		selected:
			filesStore.shareFilter.scope.public &&
			!filesStore.shareFilter.scope.smb &&
			!filesStore.shareFilter.scope.internal,
		isAll: false,
		isDefault: false
	},
	{
		value: ShareType.SMB,
		label: shareTypeStr(ShareType.SMB),
		selected:
			!filesStore.shareFilter.scope.public &&
			filesStore.shareFilter.scope.smb &&
			!filesStore.shareFilter.scope.internal,
		isAll: false,
		isDefault: false
	},
	{
		value: ShareType.INTERNAL,
		label: shareTypeStr(ShareType.INTERNAL),
		selected:
			!filesStore.shareFilter.scope.public &&
			!filesStore.shareFilter.scope.smb &&
			filesStore.shareFilter.scope.internal,
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
		selected:
			filesStore.shareFilter.permission.manage &&
			!filesStore.shareFilter.permission.edit &&
			!filesStore.shareFilter.permission.view,
		isAll: false,
		isDefault: false
	},
	{
		value: SharePermission.Edit,
		label: sharePermissionStr(SharePermission.Edit),
		selected:
			!filesStore.shareFilter.permission.manage &&
			filesStore.shareFilter.permission.edit &&
			!filesStore.shareFilter.permission.view,
		isAll: false,
		isDefault: false
	},
	{
		value: SharePermission.View,
		label: sharePermissionStr(SharePermission.View),
		selected:
			!filesStore.shareFilter.permission.manage &&
			!filesStore.shareFilter.permission.edit &&
			filesStore.shareFilter.permission.view,
		isAll: false,
		isDefault: false
	}
]);

const expirTimeSelect = ref({
	expand: false,
	values: [
		{
			value: ExoirationTime.all,
			label: t('my.all'),
			selected: filesStore.shareFilter.expire == ExoirationTime.all,
			isAll: true,
			isDefault: false
		},
		{
			value: ExoirationTime.within1days,
			label: t('files.With in 1 day'),
			selected: filesStore.shareFilter.expire == ExoirationTime.within1days,
			isAll: false,
			isDefault: false
		},
		{
			value: ExoirationTime.within7days,
			label: t('files.With in {number} days', {
				number: 7
			}),
			selected: filesStore.shareFilter.expire == ExoirationTime.within7days,
			isAll: false,
			isDefault: false
		},
		{
			value: ExoirationTime.within30days,
			label: t('files.With in {number} days', {
				number: 30
			}),
			selected: filesStore.shareFilter.expire == ExoirationTime.within30days,
			isAll: false,
			isDefault: false
		},
		{
			value: ExoirationTime.within1year,
			label: t('files.With in 1 year'),
			selected: filesStore.shareFilter.expire == ExoirationTime.within1year,
			isAll: false,
			isDefault: false
		},
		{
			value: ExoirationTime.over1year,
			label: t('files.Over 1 year'),
			selected: filesStore.shareFilter.expire == ExoirationTime.over1year,
			isAll: false,
			isDefault: false
		}
	]
});

const resetall = () => {
	filesStore.resetShareFilter();
	filesStore.setBrowserUrl(route.fullPath, DriveType.Share);
	CustomRef.value.onDialogOK();
};

const confirmFilter = () => {
	filesStore.shareFilter.shared = {
		byMe: shared.value[0].selected ? true : shared.value[1].selected,
		withMe: shared.value[0].selected ? true : shared.value[2].selected
	};

	filesStore.shareFilter.owner = owner.value.values
		.filter((e) => e.selected && !e.isAll)
		.map((e) => e.label);

	filesStore.shareFilter.scope = {
		public: scope.value[0].selected ? true : scope.value[1].selected,
		smb: scope.value[0].selected ? true : scope.value[2].selected,
		internal: scope.value[0].selected ? true : scope.value[3].selected
	};

	filesStore.shareFilter.permission = {
		manage: permission.value[0].selected ? true : permission.value[1].selected,
		edit: permission.value[0].selected ? true : permission.value[2].selected,
		view: permission.value[0].selected ? true : permission.value[3].selected
	};

	const expireSelect = expirTimeSelect.value.values.find((e) => e.selected);

	if (expireSelect) {
		filesStore.shareFilter.expire = expireSelect.value;
	}

	filesStore.setBrowserUrl(route.fullPath, DriveType.Share);
	CustomRef.value.onDialogOK();
};
</script>

<style scoped lang="scss">
.footer {
	width: 100%;
	height: 88px;
	.btn-item {
		height: 48px;
	}
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
</style>
