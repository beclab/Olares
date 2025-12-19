<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="
			isMobile
				? t('share_repo')
				: `${t('share_repo')}: ${shareRepoInfo.repo_name}`
		"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		@onSubmit="onDialogOk"
		@onCancel="onDialogCancel"
		@onHide="onDialogCancel"
	>
		<div class="q-dialog-plugin-web">
			<q-card-section class="q-px-none">
				<div class="text-catipon text-ink-3">{{ t('add_user') }}</div>
				<q-item dense class="row justify-between q-mt-none q-px-none">
					<q-select
						class="multiple"
						dense
						behavior="menu"
						v-model="userModel"
						multiple
						:options="userOptions"
						borderless
						emit-value
						map-options
						option-value="email"
						option-label="name"
						placeholder="Select User"
						color="yellow"
						dropdown-icon="sym_r_expand_more"
						:style="{ width: '200px' }"
					>
						<template v-slot:option="{ itemProps, opt }">
							<q-item
								v-bind="itemProps"
								:style="{
									'pointer-events': checkShared(opt.name) ? 'none' : 'auto'
								}"
							>
								<q-item-section
									class="text-ink-1"
									style="word-break: normal; white-space: nowrap"
								>
									<span>{{ opt.name }}</span>
								</q-item-section>
								<q-item-section side>
									<span class="added text-body3" v-if="checkShared(opt.name)">
										{{ t('added') }}
									</span>
								</q-item-section>
							</q-item>
						</template>
					</q-select>
					<span
						class="adduser text-body3"
						:class="!userModel || userModel.length <= 0 ? 'adduserdisable' : ''"
						:style="{
							'pointer-events':
								!userModel || userModel.length <= 0 ? 'none' : 'auto'
						}"
						@click="submit"
					>
						{{ t('add_user') }}
					</span>
				</q-item>
			</q-card-section>

			<q-card-section class="q-pt-none q-px-none">
				<q-item
					dense
					class="row items=center justify-between q-mt-none q-px-none text-ink-3"
				>
					<span class="text-catipon text-grey-7">{{ t('user') }}</span>
					<span class="text-catipon text-grey-7">{{
						t('permission.index')
					}}</span>
				</q-item>

				<q-item
					dense
					class="row q-pa-none items-center justify-between text-ink-2"
					v-for="item in menuStore.sharedItems"
					:key="item.user_info.name"
				>
					<q-item-section class="col-6 row items-start q-py-none">
						{{ item.user_info.nickname }}
					</q-item-section>
					<div class="col-6 row items-center justify-end q-py-none">
						<q-select
							dense
							behavior="menu"
							v-model="item.permission"
							:options="permissionOption"
							borderless
							emit-value
							map-options
							dropdown-icon="sym_r_expand_more"
							@update:model-value="
								updateItem(item.user_info.name, item.permission)
							"
						>
							<template v-slot:selected>
								<span class="text-ink-2">
									{{
										item.permission.replace(/\s*/g, '') === 'rw'
											? 'Read-Write'
											: 'Read-only'
									}}
								</span>
							</template>

							<template v-slot:option="scope">
								<q-item v-bind="scope.itemProps">
									<q-item-section>
										<q-item-label>{{ scope.opt.label }}</q-item-label>
										<q-item-label caption>{{
											scope.opt.description
										}}</q-item-label>
									</q-item-section>
								</q-item>
							</template>
						</q-select>
						<q-icon
							class="cursor-pointer q-ml-sm"
							name="sym_r_delete"
							size="22px"
							@click="deleteShareItem(item.user_info.name)"
						></q-icon>
					</div>
				</q-item>
			</q-card-section>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { useQuasar } from 'quasar';
import { shareToUser } from '../../../api';
import { useDataStore } from '../../../stores/data';
import { useMenuStore } from '../../../stores/files-menu';
import { useFilesStore } from '../../../stores/files';

import { useI18n } from 'vue-i18n';

const $q = useQuasar();
const store = useDataStore();
const menuStore = useMenuStore();
const filesStore = useFilesStore();

const { t } = useI18n();

const userModel = ref();
const userOptions = ref();
const primaryModel = ref('rw');
const permissionOption = ref([
	{
		label: 'Read-Write',
		value: 'rw',
		description: t('files.user_can_read_write_upload_download_and_sync_files')
	},
	{
		label: 'Read-only',
		value: 'r',
		description: t('files.user_can_read_download_and_sync_files')
	}
]);
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);
const shareRepoInfo = ref(menuStore.shareRepoInfo);

const submit = async () => {
	try {
		await shareToUser().setSharedItems(userModel.value, primaryModel.value);
		await menuStore.listSharedItems();
		filesStore.getMenu();

		userModel.value = null;
	} catch (error) {
		console.error('error', error);
	}
};

const updateItem = async (name: string, permission: string) => {
	try {
		await shareToUser().updateItem(name, permission);
		await menuStore.listSharedItems();
		filesStore.getMenu();
	} catch (error) {
		console.error('error', error);
	}
};

const deleteShareItem = async (name: string) => {
	try {
		await shareToUser().deleteItem(name);
		await menuStore.listSharedItems();
		filesStore.getMenu();
	} catch (error) {
		console.error('error', error);
	}
};

const checkShared = (name: string) => {
	const isShared = menuStore.sharedItems.find(
		(item) => item.user_info.nickname === name
	);
	if (isShared) {
		return true;
	} else {
		return false;
	}
};

const onDialogCancel = () => {
	store.closeHovers();
};

const onDialogOk = () => {
	store.closeHovers();
};

onMounted(async () => {
	await menuStore.listSharedItems();
	await menuStore.getUserList();
	userOptions.value = menuStore.userList;
});
</script>

<style lang="scss" scoped>
.q-dialog-plugin-web {
	border-radius: 12px;
	padding: 0;

	.multiple {
		flex: 1;
		border: 1px solid $input-stroke;
		border-radius: 8px;
		overflow: hidden;
		padding-left: 8px;
	}

	.adduser {
		line-height: 40px;
		padding: 0 10px;
		margin-left: 10px;
		border-radius: 8px;
		border: 1px solid $yellow-disabled;
		word-break: normal;
		white-space: nowrap;
		color: $ink-1;

		cursor: pointer;
		&:hover {
			background-color: $background-hover;
		}
		&.adduserdisable {
			color: $ink-3;
			border: 1px solid $yellow-disabled;
		}
	}
}

.added {
	padding: 4px 8px;
	border-radius: 4px;
	border: 1px solid $separator;
}
</style>
