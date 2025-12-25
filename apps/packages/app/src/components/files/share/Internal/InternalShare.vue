<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files_popup_menu.Share to Internal')"
		:skip="false"
		:ok="ok"
		:cancel="cancel"
		:persistent="true"
		size="medium"
		:cancelDismiss="currentStep == ShareAddStep.BASE"
		:okDisabled="onDisabled"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
		:disableCancelFucus="cancelFocusDisable"
	>
		<template v-slot:header v-if="currentStep == ShareAddStep.INVITE_PEOPLE">
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
		<template
			v-slot:header
			v-else-if="currentStep == ShareAddStep.LINK_GENERATE"
		>
			<share-back
				:title="t('files.Link Settings')"
				@back="onCancel"
				platform="web"
		/></template>

		<div class="dialog-desc" v-if="filesStore.users">
			<template v-if="currentStep == ShareAddStep.BASE && filesStore.users">
				<div class="text-subtitle1 text-ink-2">
					{{ t('files.Invite people') }}
				</div>
				<ShareUserSelect
					:isTextarea="false"
					:isReadOnly="true"
					:hintText="t('files.Search for a user, group')"
					@click="currentStep = ShareAddStep.INVITE_PEOPLE"
				/>
				<div class="row items-center justify-between user-permission q-mt-xl">
					<div class="text-subtitle1 text-ink-1">
						{{ t('files.User permissions Settings') }}
					</div>
					<div
						class="users-bg q-px-xs row items-center justify-between"
						@click="currentStep = ShareAddStep.USER_PERMISSION"
					>
						<div class="q-px-xs row items-center">
							<q-avatar :size="`24px`" class="">
								<TerminusAvatar
									:info="{
										terminusName: filesStore.users?.olaresId
									}"
									:size="24"
								/>
							</q-avatar>
							<template v-if="selectedUsers.length > 1">
								<div
									style="height: 20px; width: 1px"
									class="bg-separator q-ml-sm"
								></div>
								<q-avatar :size="`24px`" class="q-ml-sm">
									<TerminusAvatar
										:info="{
											terminusName: selectedUsers[1].olaresId
										}"
										:size="24"
									/>
								</q-avatar>
								<q-avatar
									:size="`24px`"
									class="q-ml-sm"
									v-if="selectedUsers.length > 2"
								>
									<TerminusAvatar
										:info="{
											terminusName: selectedUsers[2].olaresId
										}"
										:size="24"
									/>
								</q-avatar>
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
			<template v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE">
				<ShareUserSelect
					:isTextarea="false"
					:hintText="t('files.Search for a user, group')"
					@click="currentStep = ShareAddStep.INVITE_PEOPLE"
					:users="users"
				/>
			</template>
			<template v-else-if="currentStep == ShareAddStep.USER_PERMISSION">
				<ShareUserPermissionSetting
					:currentUser="filesStore.users.owner"
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
import {
	FilesIdType,
	useFilesStore,
	ShareItemUser,
	ShareMember
} from '../../../../stores/files';
import { useDataStore } from '../../../../stores/data';
import ShareUserSelect from '../ShareUserSelect.vue';
import ShareBack from '../ShareBack.vue';
import ShareUserPermissionSetting from './ShareUserPermissionSetting.vue';
import * as filesUtil from '../../../../api/files/v2/common/utils';
import share from '../../../../api/files/v2/common/share';
import { ShareType, SharePermission } from 'src/utils/interface/share';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { ShareAddStep, useInternalShare } from './internal';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	}
});

const {
	internalShareId,
	initUsers,
	createInternalShare,
	selectedUsers,
	CustomRef,
	cancelFocusDisable,
	createMembers,
	currentStep,
	updateShareMember
} = useInternalShare(props.origin_id);

const store = useDataStore();

const { t } = useI18n();

const filesStore = useFilesStore();

const users = ref([] as { name: string; selected: boolean }[]);

const onCancel = () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		currentStep.value = ShareAddStep.BASE;
		users.value.forEach((e) => (e.selected = false));
	} else if (currentStep.value == ShareAddStep.BASE) {
		store.closeHovers();
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	}
};

const cancel = computed(() => {
	if (currentStep.value == ShareAddStep.BASE) {
		return t('cancel');
	}
	if (
		currentStep.value == ShareAddStep.INVITE_PEOPLE ||
		currentStep.value == ShareAddStep.USER_PERMISSION
	) {
		return t('cancel');
	}
	return false;
});

const ok = computed(() => {
	if (currentStep.value == ShareAddStep.BASE) {
		return t('confirm');
	}
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		return t('invite');
	}
	if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		return t('submit');
	}

	return false;
});

const onDisabled = computed(() => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		return users.value.filter((e) => e.selected).length == 0;
	}
	if (currentStep.value == ShareAddStep.BASE) {
		return false;
	}

	if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		return selectedUsers.value.length == 0;
	}

	return false;
});

const onSubmit = async () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		const leftUser = [] as { name: string; selected: boolean }[];
		users.value.forEach((e) => {
			if (e.selected) {
				const user = filesStore.users?.users.find((l) => l.name == e.name);
				if (user) {
					selectedUsers.value.push({
						...user,
						isOwner: false,
						permission: SharePermission.View,
						editingPermission: SharePermission.View
					});
				}
			} else {
				leftUser.push(e);
			}
		});
		users.value = leftUser;
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.BASE) {
		try {
			if (!!internalShareId.value) {
				await updateShareMember();
				notifySuccess(t('success'));
			} else {
				const shareId = await createInternalShare();
				if (!shareId) {
					return;
				}
				await createMembers(shareId);
				notifySuccess(t('success'));
			}

			CustomRef.value.onDialogOK();
			store.closeHovers();
		} catch (error) {
			notifyFailed(error.message);
		}
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	}
};

const deleteItem = (item: ShareItemUser) => {
	selectedUsers.value = selectedUsers.value.filter((e) => e.name != item.name);
	users.value.push({
		name: item.name,
		selected: false
	});
};

onMounted(async () => {
	const { owner, members } = await initUsers();
	if (filesStore.users) {
		filesStore.users.users.forEach((e) => {
			if (members.find((m) => m.share_member == e.name)) {
				selectedUsers.value.push({
					...e,
					permission: members.find((m) => m.share_member == e.name)!.permission,
					isOwner: e.name == owner,
					editingPermission: members.find((m) => m.share_member == e.name)!
						.permission
				});
			} else if (e.name == filesStore.users!.owner || owner == e.name) {
				selectedUsers.value.push({
					...e,
					permission: SharePermission.ADMIN,
					isOwner: true,
					editingPermission: SharePermission.ADMIN
				});
			} else {
				users.value.push({
					name: e.name,
					selected: false
				});
			}
		});
	}
});
</script>

<style lang="scss" scoped>
.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.isSelectedMode {
		opacity: 0.5;
		pointer-events: none;
	}

	.q-tab {
		padding: 0 0px;
		margin-right: 8px;
	}

	.internal-tab {
		.user-permission {
			width: 100%;
			height: 40px;

			.users-bg {
				height: 32px;
				border-radius: 16px;
				background: $background-3;
				min-width: 80px;
			}

			.users-bg:hover {
				background-color: $background-4;
			}
		}
	}

	.generate-password {
		border: 1px solid $light-blue-default;
		padding: 0px 12px;
		border-radius: 8px;
		height: 40px;
		cursor: pointer;

		&:hover {
			background-color: $background-3;
		}
	}

	.create-link {
		border: 1px solid $light-blue-default;
		width: 100%;
		height: 40px;
		border-radius: 8px;
		cursor: pointer;
		&:hover {
			background-color: $background-3;
		}
		&.operate-disabled {
			opacity: 0.5;
		}
	}
}

.footerMore {
	width: 100px;
	position: absolute;
	left: 20px;
	border-radius: 8px;
	font-weight: 500;
	font-size: 16px;
	padding: 8px 0;
	line-height: 24px;
}
</style>
