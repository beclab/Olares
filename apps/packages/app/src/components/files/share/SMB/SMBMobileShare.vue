<template>
	<bt-custom-dialog
		ref="CustomRef"
		:skip="false"
		:ok="ok"
		:cancel="cancel"
		:persistent="true"
		size="medium"
		platform="mobile"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
		position="bottom"
	>
		<template v-slot:header v-if="currentStep == ShareAddStep.BASE">
			<share-back :title="t('files_popup_menu.Share to SMB')" @back="onBack">
				<template v-slot:end>
					<q-icon
						name="sym_r_close"
						color="ink-2"
						size="24px"
						@click="onClose"
					/>
				</template>
			</share-back>
		</template>
		<template
			v-slot:header
			v-else-if="
				currentStep == ShareAddStep.ADD_ACCOUNT ||
				currentStep == ShareAddStep.INVITE_PEOPLE ||
				currentStep == ShareAddStep.USER_PERMISSION ||
				currentStep == ShareAddStep.INVITE_PEOPLE_PERMISSION
			"
		>
			<share-back
				:title="
					currentStep == ShareAddStep.ADD_ACCOUNT
						? t('files.Add user accounts')
						: currentStep == ShareAddStep.INVITE_PEOPLE
						? t('files.Invite people')
						: t('files.User permissions Settings')
				"
				@back="onCancel"
			>
				<template v-slot:end>
					<q-icon
						name="sym_r_close"
						color="ink-2"
						size="24px"
						@click="onClose"
					/>
				</template>
			</share-back>
		</template>
		<div
			class="dialog-desc"
			:class="{
				'set-height':
					currentStep == ShareAddStep.BASE ||
					currentStep == ShareAddStep.ADD_ACCOUNT
			}"
		>
			<template v-if="currentStep == ShareAddStep.BASE">
				<template v-if="!shareResult">
					<div class="text-ink-2 text-subtitle1">
						{{ t('public') }}
					</div>
					<div class="row items-center q-mt-sm" style="height: 36px">
						<bt-check-box-component
							class="col"
							:model-value="smbPublic == true"
							:label="'Yes'"
							@update:modelValue="smbPublic = true"
						/>
						<bt-check-box-component
							class="col"
							:model-value="smbPublic == false"
							:label="'No'"
							@update:modelValue="smbPublic = false"
						/>
					</div>
					<template v-if="!smbPublic">
						<terminus-item
							:show-board="false"
							iconName="sym_r_account_circle"
							:whole-picture-size="24"
							:icon-size="24"
							:item-height="48"
							:padding-left="0"
							@click="currentStep = ShareAddStep.ADD_ACCOUNT"
						>
							<template v-slot:title>
								<div class="text-subtitle2 text-ink-2">
									{{ t('files.Add user accounts') }}
								</div>
							</template>
							<template v-slot:side>
								<q-icon
									name="sym_r_keyboard_arrow_right"
									size="20px"
									color="ink-3"
								/>
							</template>
						</terminus-item>

						<terminus-item
							:show-board="false"
							iconName="sym_r_person_add"
							:whole-picture-size="24"
							:icon-size="24"
							:item-height="48"
							:padding-left="0"
							@click="currentStep = ShareAddStep.INVITE_PEOPLE"
						>
							<template v-slot:title>
								<div class="text-subtitle2 text-ink-2">
									{{ t('files.Invite people') }}
								</div>
							</template>
							<template v-slot:side>
								<q-icon
									name="sym_r_keyboard_arrow_right"
									size="20px"
									color="ink-3"
								/>
							</template>
						</terminus-item>
						<terminus-item
							:show-board="false"
							iconName="sym_r_manage_accounts"
							:whole-picture-size="24"
							:icon-size="24"
							:item-height="48"
							:padding-left="0"
							style="margin-bottom: 100px"
							@click="currentStep = ShareAddStep.USER_PERMISSION"
							side-style="width: 50%; flex: 1"
						>
							<template v-slot:title>
								<div class="text-subtitle2 text-ink-2">
									{{ t('files.User permissions Settings') }}
								</div>
							</template>
							<template v-slot:side>
								<div
									class="row items-center justify-end full-width"
									v-if="selectedUsers.length > 0"
								>
									<StackedAvatars
										:avatarsLength="selectedUsers.length"
										:max-visible="3"
									>
										<template v-slot:content="props">
											<SMBUserIcon
												:name="selectedUsers[props.data - 1].name"
												:size="20"
												:inner-size="16"
											/>
										</template>
									</StackedAvatars>
									<q-icon
										name="sym_r_keyboard_arrow_right"
										size="20px"
										color="ink-3"
									/>
								</div>
								<q-icon
									v-else
									name="sym_r_keyboard_arrow_right"
									size="20px"
									color="ink-3"
								/>
							</template>
						</terminus-item>
					</template>
				</template>
				<template v-else>
					<div class="share-link row items-center q-pa-md" @click="copy">
						<div class="share-info text-body2 text-ink-3">
							<div class="row justify-start">
								<!-- <div class="title">{{ t('Link') }}</div> -->
								<div class="detail text-ink-2">
									{{ getShareLink }}
								</div>
							</div>
						</div>
					</div>
					<template v-if="selectedUsers.length > 0 && !smbPublic">
						<div
							class="share-link q-pa-md q-mt-sm row items-center text-ink-3"
							v-for="user in selectedUsers"
							:key="user.id"
						>
							<div class="row items-center justify-start share-info">
								<div class="title">{{ t('accounts') }}</div>
								<div class="detail text-ink-2">
									{{ user.name }}
								</div>
							</div>
							<div class="row items-center justify-start share-info q-mt-md">
								<div class="title">{{ t('password') }}</div>
								<div class="text-ink-2">
									{{ user.pwdDisplay ? user.password : '······' }}
								</div>
								<q-icon
									class="q-ml-md"
									color="blue-default"
									size="16px"
									:name="
										user.pwdDisplay
											? 'sym_r_visibility_off'
											: 'sym_r_visibility'
									"
									@click="user.pwdDisplay = !user.pwdDisplay"
								/>
							</div>
						</div>
					</template>
				</template>
			</template>
			<template v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE">
				<ShareMobileUserSelect
					:isTextarea="false"
					:hintText="t('files.Search for a user, group')"
					@click="currentStep = ShareAddStep.INVITE_PEOPLE"
					:users="smbMobileUsers"
				>
					<template v-slot:list-avatar="props">
						<SMBUserIcon :name="props.user.name" />
					</template>
					<template v-slot:select-avatar="props">
						<SMBUserIcon :name="props.user.name" :size="24" />
					</template>
				</ShareMobileUserSelect>
			</template>
			<template v-else-if="currentStep == ShareAddStep.ADD_ACCOUNT">
				<AddAccount
					:account="smbAccount"
					:password="smbPassword"
					@update:account="
						(value) => {
							smbAccount = value;
						}
					"
					@update:password="
						(value) => {
							smbPassword = value;
						}
					"
				/>
			</template>
			<template
				v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE_PERMISSION"
			>
				<ShareMobilePermissionSetting
					:users="smbMobileUsers.filter((e) => e.selected)"
					@edit-permission="editPermission"
				>
					<template v-slot:list-avatar="props">
						<SMBUserIcon :name="props.user.name" />
					</template>
				</ShareMobilePermissionSetting>
			</template>
			<template v-else-if="currentStep == ShareAddStep.USER_PERMISSION">
				<ShareMobilePermissionSetting
					:users="selectedUsers"
					@edit-permission="editPermission"
				>
					<template v-slot:list-avatar="props">
						<SMBUserIcon :name="props.user.name" />
					</template>
				</ShareMobilePermissionSetting>
			</template>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted } from 'vue';
import { FilesIdType, useFilesStore } from '../../../../stores/files';
import { useDataStore } from '../../../../stores/data';
import ShareBack from '../ShareBack.vue';
import { busEmit } from 'src/utils/bus';
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import { useSMBShare, ShareAddStep } from './smb';
import TerminusItem from 'src/components/common/TerminusItem.vue';
import SMBUserIcon from './SMBUserIcon.vue';
import AddAccount from './AddAccount.vue';
import ShareMobileUserSelect from '../ShareMobileUserSelect.vue';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import ShareMobilePermissionSetting from '../ShareMobilePermissionSetting.vue';
import StackedAvatars from '../StackedAvatars.vue';
import { useQuasar } from 'quasar';
import { SharePermission } from 'src/utils/interface/share';
import ShareMobileEditPermissionDialog from './ShareMobileEditPermissionDialog.vue';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const {
	smbPassword,
	shareResult,
	smbPublic,
	createSMBShare,
	getShareLink,
	copy,
	currentStep,
	selectedUsers,
	smbAccount,
	smbMobileUsers,
	initSMBUsers,
	createSMBUser,
	totalSMBUsers,
	initShareInfo
} = useSMBShare(props.origin_id);

const store = useDataStore();
const { t } = useI18n();
const filesStore = useFilesStore();

const CustomRef = ref();

const cancel = computed(() => {
	return false;
});

const ok = computed(() => {
	return t('confirm');
});

const onSubmit = async () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		currentStep.value = ShareAddStep.INVITE_PEOPLE_PERMISSION;
	} else if (currentStep.value == ShareAddStep.INVITE_PEOPLE_PERMISSION) {
		smbMobileUsers.value.forEach((e) => {
			const selectedItemIndex = selectedUsers.value.findIndex(
				(item) => item.id == e.id
			);
			if (e.selected) {
				if (selectedItemIndex < 0) {
					const user = totalSMBUsers.value.find((l) => l.id == e.id);
					selectedUsers.value.push({
						...user!,
						permission: e.permission,
						editingPermission: e.permission,
						pwdDisplay: false
					});
				} else {
					selectedUsers.value[selectedItemIndex].permission = e.permission;
					selectedUsers.value[selectedItemIndex].editingPermission =
						e.permission;
				}
			} else {
				if (selectedItemIndex >= 0) {
					selectedUsers.value.splice(selectedItemIndex, 1);
				}
			}
		});
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.BASE) {
		if (!shareResult.value) {
			createSMBShare();
		} else {
			copy();
			onClose();
		}
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.ADD_ACCOUNT) {
		const result = await createSMBUser();
		if (result) {
			currentStep.value = ShareAddStep.BASE;
			notifySuccess(t('success'));
		}
	}
};

const onClose = () => {
	store.closeHovers();
	CustomRef.value.onDialogCancel();
};

const onBack = () => {
	const index = filesStore.selected[props.origin_id][0];
	store.closeHovers();
	busEmit('fileItemOpenOperation', index);
	CustomRef.value.onDialogCancel();
};

const onCancel = () => {
	if (
		currentStep.value == ShareAddStep.INVITE_PEOPLE ||
		currentStep.value == ShareAddStep.ADD_ACCOUNT ||
		currentStep.value == ShareAddStep.USER_PERMISSION
	) {
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.INVITE_PEOPLE_PERMISSION) {
		currentStep.value = ShareAddStep.INVITE_PEOPLE;
	} else if (currentStep.value == ShareAddStep.BASE) {
		store.closeHovers();
	}
};

const deleteItem = (user: { name: string }) => {
	selectedUsers.value = selectedUsers.value.filter((e) => e.name != user.name);
	const editUser = smbMobileUsers.value.find((e) => e.name == user.name);
	if (!editUser) {
		return;
	}
	editUser.selected = false;
};
const $q = useQuasar();

const editPermission = (user: {
	permission: SharePermission;
	name: string;
}) => {
	$q.dialog({
		component: ShareMobileEditPermissionDialog,
		componentProps: {
			user: user
		}
	}).onOk((info: { permission: SharePermission; remove: boolean }) => {
		if (info.remove) {
			deleteItem(user);
		} else {
			user.permission = info.permission;
		}
	});
};

onMounted(async () => {
	await initSMBUsers();
	await initShareInfo();
});
</script>

<style lang="scss" scoped>
.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.share-link {
		width: 100%;
		border-radius: 8px;
		background: $background-6;

		.share-info {
			width: 100%;
			.title {
				width: 64px;
				padding-top: 4px;
			}

			.detail {
				// flex: auto;
				max-width: calc(100% - 64px);
			}
		}
	}
}

.set-height {
	height: 330px;
}
</style>
