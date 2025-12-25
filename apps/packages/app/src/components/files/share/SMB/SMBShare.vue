<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files_popup_menu.Share to SMB')"
		:skip="false"
		:ok="ok"
		:cancel="cancel"
		:persistent="true"
		size="medium"
		:cancelDismiss="currentStep == ShareAddStep.BASE"
		:okDisabled="onDisabled"
		:disableCancelFucus="cancelFocusDisable"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
	>
		<template v-slot:header v-if="currentStep == ShareAddStep.ADD_ACCOUNT">
			<share-back
				:title="t('files.Add user accounts')"
				@back="onCancel"
				platform="web"
			/>
		</template>
		<template
			v-slot:header
			v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE"
		>
			<share-back
				:title="t('files.Invite people')"
				@back="onCancel"
				platform="web"
			/>
		</template>
		<template
			v-slot:header
			v-else-if="currentStep == ShareAddStep.USER_PERMISSION"
		>
			<share-back
				:title="t('files.User permissions Settings')"
				@back="onCancel"
				platform="web"
			/>
		</template>
		<div class="dialog-desc">
			<template v-if="currentStep == ShareAddStep.BASE">
				<template v-if="!shareResult">
					<div class="text-ink-2 text-subtitle1">
						{{ t('public') }}
					</div>
					<div class="row items-center q-mt-sm">
						<bt-check-box-component
							:model-value="smbPublic == true"
							:label="'Yes'"
							@update:modelValue="smbPublic = true"
						/>
						<bt-check-box-component
							:model-value="smbPublic == false"
							:label="'No'"
							class="q-ml-lg"
							@update:modelValue="smbPublic = false"
						/>
					</div>
					<template v-if="!smbPublic">
						<div
							class="text-subtitle1 text-ink-2 row items-center justify-between q-mt-sm"
							style="height: 40px"
						>
							<div>
								{{ t('files.Invite people') }}
							</div>
							<div
								class="add_account"
								@click="currentStep = ShareAddStep.ADD_ACCOUNT"
							>
								{{ t('files.Add user accounts') }}
							</div>
						</div>
						<ShareUserSelect
							:isTextarea="false"
							:isReadOnly="true"
							:hintText="t('files.Search for a user, group')"
							@click="currentStep = ShareAddStep.INVITE_PEOPLE"
						/>
						<div
							class="row items-center justify-between user-permission q-mt-lg"
						>
							<div class="text-subtitle1 text-ink-2">
								{{ t('files.User permissions Settings') }}
							</div>
							<div
								class="users-bg q-px-xs row items-center justify-between"
								@click="currentStep = ShareAddStep.USER_PERMISSION"
							>
								<div class="q-px-xs row items-center">
									<template v-if="selectedUsers.length > 0">
										<SMBUserIcon :name="selectedUsers[0].name" />
										<SMBUserIcon
											:name="selectedUsers[1].name"
											class="q-ml-xs"
											v-if="selectedUsers.length > 1"
										/>
										<SMBUserIcon
											:name="selectedUsers[2].name"
											class="q-ml-xs"
											v-if="selectedUsers.length > 2"
										/>
									</template>
									<div
										class="q-px-sm text-body3 tex-ink-2"
										v-if="selectedUsers.length > 3"
									>
										+{{ selectedUsers.length - 3 }}
									</div>
								</div>
								<q-icon name="sym_r_keyboard_arrow_right" size="20px" />
							</div>
						</div>
					</template>
				</template>
				<template v-else>
					<div
						class="share-link row items-center q-pa-md text-ink-3"
						@click="copy"
					>
						<div class="share-info text-body2">
							<div class="row items-center justify-start">
								<div class="title">{{ t('Link') }}</div>
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
				<ShareUserSelect
					:isTextarea="false"
					:hintText="t('files.Search for a user, group')"
					@click="currentStep = ShareAddStep.INVITE_PEOPLE"
					:users="smbUsers"
				/>
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
			<template v-else-if="currentStep == ShareAddStep.USER_PERMISSION">
				<ShareUserPermissionSetting
					:users="selectedUsers"
					@delete="deleteItem"
					@update-menu="(status: boolean)=> {
						cancelFocusDisable = status;
					}"
				/>
			</template>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted } from 'vue';
import { FilesIdType } from '../../../../stores/files';
import { useDataStore } from '../../../../stores/data';
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import { ShareAddStep, useSMBShare } from './smb';
import AddAccount from './AddAccount.vue';
import ShareBack from '../ShareBack.vue';
import ShareUserSelect from '../ShareUserSelect.vue';
import { SharePermission } from 'src/utils/interface/share';
import SMBUserIcon from './SMBUserIcon.vue';
import { notifySuccess } from 'src/utils/notifyRedefinedUtil';
import ShareUserPermissionSetting from './ShareUserPermissionSetting.vue';

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
	initSMBUsers,
	smbUsers,
	selectedUsers,
	totalSMBUsers,
	smbAccount,
	createSMBUser,
	deleteItem,
	currentStep,
	initShareInfo
	// onCancel
} = useSMBShare(props.origin_id);

const store = useDataStore();
const { t } = useI18n();
const cancelFocusDisable = ref(false);

const cancel = computed(() => {
	if (shareResult.value == undefined) {
		return t('cancel');
	}
	return false;
});

const ok = computed(() => {
	if (currentStep.value == ShareAddStep.BASE) {
		if (shareResult.value) {
			return t('files.Copy link');
		}
	}
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		return t('invite');
	}
	if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		return t('submit');
	}
	return t('confirm');
});

const CustomRef = ref();

const onSubmit = async () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		const leftUser = [] as { name: string; selected: boolean; id: string }[];
		smbUsers.value.forEach((e) => {
			if (e.selected) {
				const user = totalSMBUsers.value.find((l) => l.id == e.id);
				if (user) {
					selectedUsers.value.push({
						...user,
						permission: SharePermission.View,
						editingPermission: SharePermission.View,
						pwdDisplay: false
					});
				}
			} else {
				leftUser.push(e);
			}
		});
		smbUsers.value = leftUser;
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.ADD_ACCOUNT) {
		const result = await createSMBUser();
		if (result) {
			currentStep.value = ShareAddStep.BASE;
			notifySuccess(t('success'));
		}
	} else if (currentStep.value == ShareAddStep.BASE) {
		if (!shareResult.value) {
			createSMBShare();
		} else {
			copy();
			store.closeHovers();
		}
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	}
};

const onCancel = () => {
	if (currentStep.value == ShareAddStep.ADD_ACCOUNT) {
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.BASE) {
		store.closeHovers();
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		currentStep.value = ShareAddStep.BASE;
	}
};

const onDisabled = computed(() => {
	if (currentStep.value == ShareAddStep.ADD_ACCOUNT) {
		return !!smbPassword.value && !!smbAccount.value ? false : true;
	}

	return false;
});

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
		min-height: 60px;
		border-radius: 8px;
		background: $background-6;

		.share-info {
			width: 100%;
			.title {
				width: 64px;
			}

			.detail {
				max-width: calc(100% - 64px);
			}
		}
	}

	.add_account {
		cursor: pointer;
		color: $light-blue-default;
		&:hover {
			color: $blue-default;
		}
	}
}
</style>
