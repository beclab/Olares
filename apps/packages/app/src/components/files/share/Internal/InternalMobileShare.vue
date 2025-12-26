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
		position="bottom"
		platform="mobile"
		@onCancel="onCancel"
		@onSubmit="onSubmit"
	>
		<template v-slot:header v-if="currentStep == ShareAddStep.BASE">
			<share-back
				:title="t('files_popup_menu.Share to Internal')"
				@back="onBack"
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
		<template
			v-slot:header
			v-else-if="
				currentStep == ShareAddStep.INVITE_PEOPLE ||
				currentStep == ShareAddStep.INVITE_PEOPLE_PERMISSION
			"
		>
			<share-back :title="t('files.Invite people')" @back="onCancel">
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
			v-else-if="currentStep == ShareAddStep.USER_PERMISSION"
		>
			<share-back
				:title="t('files.User permissions Settings')"
				@back="onCancel"
			>
				<template v-slot:end>
					<q-icon
						name="sym_r_close"
						color="ink-2"
						size="24px"
						@click="onCancel"
					/>
				</template>
			</share-back>
		</template>
		<template
			v-slot:header
			v-else-if="currentStep == ShareAddStep.LINK_GENERATE"
		>
			<share-back :title="t('files.Link Settings')" @back="onCancel"
		/></template>

		<div class="dialog-desc" v-if="filesStore.users">
			<template v-if="currentStep == ShareAddStep.BASE && filesStore.users">
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
									<q-avatar size="20px">
										<TerminusAvatar
											:info="{
												terminusName: selectedUsers[props.data - 1].olaresId
											}"
											:size="20"
										/>
									</q-avatar>
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
			<template v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE">
				<ShareMobileUserSelect
					:isTextarea="false"
					:hintText="t('files.Search for a user, group')"
					@click="currentStep = ShareAddStep.INVITE_PEOPLE"
					:users="users"
				>
					<template v-slot:list-avatar="props">
						<q-avatar :size="`32px`" class="">
							<TerminusAvatar
								:info="{
									terminusName: props.user.olaresId
								}"
								:size="32"
							/>
						</q-avatar>
					</template>
					<template v-slot:select-avatar="props">
						<q-avatar :size="`24px`" class="">
							<TerminusAvatar
								:info="{
									terminusName: props.user.olaresId
								}"
								:size="32"
							/>
						</q-avatar>
					</template>
				</ShareMobileUserSelect>
			</template>
			<template
				v-else-if="currentStep == ShareAddStep.INVITE_PEOPLE_PERMISSION"
			>
				<ShareMobilePermissionSetting
					:users="users.filter((e) => e.selected)"
					:currentUser="filesStore.users.owner"
					@edit-permission="editPermission"
				>
					<template v-slot:list-avatar="props">
						<q-avatar :size="`32px`" class="">
							<TerminusAvatar
								:info="{
									terminusName: props.user.olaresId
								}"
								:size="32"
							/>
						</q-avatar>
					</template>
				</ShareMobilePermissionSetting>
			</template>
			<template v-else-if="currentStep == ShareAddStep.USER_PERMISSION">
				<template
					v-for="item in selectedUsers.filter((e) => e.isOwner)"
					:key="item.name"
				>
					<div
						class="row items-center justify-between"
						style="width: 100%; height: 56px"
					>
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
								v-if="item.role == OLARES_ROLE.OWNER"
								class="owner bg-background-3 text-light-blue-default text-subtitle3 q-ml-md row items-center q-px-sm"
							>
								{{ t('owner') }}
							</div>
						</div>
						<div class="">
							{{ t('admin') }}
						</div>
					</div>
					<q-separator class="" />
				</template>
				<ShareMobilePermissionSetting
					:users="selectedUsers.filter((e) => e.isOwner == false)"
					:currentUser="filesStore.users.owner"
					@edit-permission="editPermission"
				>
					<template v-slot:list-avatar="props">
						<q-avatar :size="`32px`" class="">
							<TerminusAvatar
								:info="{
									terminusName: props.user.olaresId
								}"
								:size="32"
							/>
						</q-avatar>
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

import { SharePermission } from 'src/utils/interface/share';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { busEmit } from 'src/utils/bus';
import TerminusItem from 'src/components/common/TerminusItem.vue';
import ShareMobileUserSelect from '../ShareMobileUserSelect.vue';
import ShareMobilePermissionSetting from '../ShareMobilePermissionSetting.vue';
import { OLARES_ROLE } from 'src/constant';
import StackedAvatars from '../StackedAvatars.vue';
import ShareMobileEditPermissionDialog from './ShareMobileEditPermissionDialog.vue';
import { useQuasar } from 'quasar';
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
	createMembers,
	currentStep
} = useInternalShare(props.origin_id);

const store = useDataStore();

const $q = useQuasar();

const { t } = useI18n();

const filesStore = useFilesStore();

const users = ref(
	[] as {
		name: string;
		selected: boolean;
		permission: SharePermission;
		olaresId: string;
	}[]
);

const onCancel = () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		currentStep.value = ShareAddStep.BASE;
	} else if (currentStep.value == ShareAddStep.INVITE_PEOPLE_PERMISSION) {
		currentStep.value = ShareAddStep.INVITE_PEOPLE;
	} else if (currentStep.value == ShareAddStep.BASE) {
		store.closeHovers();
	} else if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		currentStep.value = ShareAddStep.BASE;
	}
};

const onBack = () => {
	const index = filesStore.selected[props.origin_id][0];
	store.closeHovers();
	busEmit('fileItemOpenOperation', index);
	CustomRef.value.onDialogCancel();
};

const onClose = () => {
	store.closeHovers();
	CustomRef.value.onDialogCancel();
};

const cancel = computed(() => {
	if (
		currentStep.value == ShareAddStep.BASE ||
		currentStep.value == ShareAddStep.INVITE_PEOPLE
	) {
		return false;
	}
	return false;
});

const ok = computed(() => {
	if (currentStep.value == ShareAddStep.BASE) {
		return t('confirm');
	}
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		return t('next');
	}
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE_PERMISSION) {
		return t('invite');
	}
	if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		return t('confirm');
	}

	return false;
});

const onDisabled = computed(() => {
	if (currentStep.value == ShareAddStep.BASE) {
		return (
			selectedUsers.value.filter((e) => e.name !== filesStore.users?.owner)
				.length == 0
		);
	}

	if (currentStep.value == ShareAddStep.USER_PERMISSION) {
		return selectedUsers.value.length == 0;
	}

	return false;
});

const onSubmit = async () => {
	if (currentStep.value == ShareAddStep.INVITE_PEOPLE) {
		if (users.value.filter((e) => e.selected).length == 0) {
			currentStep.value = ShareAddStep.BASE;
			selectedUsers.value = selectedUsers.value.filter(
				(e) => users.value.find((item) => e.name == item.name) == undefined
			);
		} else {
			currentStep.value = ShareAddStep.INVITE_PEOPLE_PERMISSION;
		}
	} else if (currentStep.value == ShareAddStep.INVITE_PEOPLE_PERMISSION) {
		users.value.forEach((e) => {
			const selectedItemIndex = selectedUsers.value.findIndex(
				(item) => item.name == e.name
			);
			if (e.selected) {
				if (selectedItemIndex < 0) {
					const user = filesStore.users?.users.find((l) => l.name == e.name);
					selectedUsers.value.push({
						...user!,
						isOwner: false,
						permission: e.permission,
						editingPermission: e.permission
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
		try {
			if (!!internalShareId.value) {
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

const deleteSelectUser = (user: { name: string }) => {
	const editUser = users.value.find((e) => e.name == user.name);
	if (!editUser) {
		return;
	}
	editUser.selected = false;
	if (users.value.filter((e) => e.selected).length == 0) {
		currentStep.value = ShareAddStep.INVITE_PEOPLE;
	}
};

onMounted(async () => {
	const { owner, members } = await initUsers();
	if (filesStore.users) {
		filesStore.users.users.forEach((e) => {
			if (members.find((m) => m.share_member == e.name)) {
				selectedUsers.value.push({
					...e,
					permission: members.find((m) => m.share_member == e.name)!.permission,
					isOwner: owner == e.name,
					editingPermission: members.find((m) => m.share_member == e.name)!
						.permission
				});
				users.value.push({
					name: e.name,
					olaresId: e.olaresId,
					selected: true,
					permission:
						members.find((m) => m.share_member == e.name)?.permission ||
						SharePermission.View
				});
			} else if (e.name == owner || e.name == filesStore.users!.owner) {
				selectedUsers.value.push({
					...e,
					permission: SharePermission.ADMIN,
					isOwner: true,
					editingPermission: SharePermission.ADMIN
				});
			} else {
				users.value.push({
					name: e.name,
					olaresId: e.olaresId,
					selected: false,
					permission: SharePermission.View
				});
			}
		});
	}
});

const editPermission = (user: {
	permission: SharePermission;
	name: string;
	olaresId: string;
}) => {
	$q.dialog({
		component: ShareMobileEditPermissionDialog,
		componentProps: {
			user: user
		}
	}).onOk((info: { permission: SharePermission; remove: boolean }) => {
		if (info.remove) {
			deleteSelectUser(user);
		} else {
			user.permission = info.permission;
		}
	});
};
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
