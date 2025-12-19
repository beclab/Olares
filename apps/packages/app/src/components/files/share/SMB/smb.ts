import { i18n } from 'src/boot/i18n';
import { SMBPermissionUser, SMBUser, useFilesStore } from 'src/stores/files';
import {
	SharePermission,
	ShareResult,
	ShareType
} from 'src/utils/interface/share';
import { notifyFailed, notifySuccess } from 'src/utils/notifyRedefinedUtil';
import { computed, ref } from 'vue';
import share from '../../../../api/files/v2/common/share';
import { useDataStore } from 'src/stores/data';
import { generatePasword } from 'src/utils/format';
import { getApplication } from 'src/application/base';

export enum ShareAddStep {
	BASE = 'base',
	ADD_ACCOUNT = 'add_account',
	USER_PERMISSION = 'user_permission',
	INVITE_PEOPLE = 'invite_person',
	INVITE_PEOPLE_PERMISSION = 'invite_persion_permission'
}

export function useSMBShare(origin_id: number) {
	const shareResult = ref<ShareResult | undefined>(undefined);
	const shareId = ref<string | undefined>(undefined);
	const smbPublic = ref(false);
	const filesStore = useFilesStore();
	const readOnly = ref(false);
	const smbPassword = ref(generatePasword(6));
	const smbAccount = ref('');
	const store = useDataStore();

	const selectedUsers = ref([] as SMBPermissionUser[]);
	const totalSMBUsers = ref([] as SMBUser[]);
	const smbUsers = ref([] as { name: string; selected: boolean; id: string }[]);
	const smbMobileUsers = ref(
		[] as {
			name: string;
			selected: boolean;
			id: string;
			permission: SharePermission;
		}[]
	);

	const currentStep = ref(ShareAddStep.BASE);

	const copy = () => {
		getApplication()
			.copyToClipboard(getShareLink.value)
			.then(() => {
				notifySuccess(i18n.global.t('copy_successfully'));
			})
			.catch(() => {
				notifyFailed(i18n.global.t('copy_fail'));
			});
	};

	const getShareLink = computed(() => {
		if (shareResult.value && shareResult.value.smb_link) {
			return shareResult.value.smb_link;
		}
		return '';
	});

	const getShareAccount = computed(() => {
		if (shareResult.value && shareResult.value.smb_user) {
			return shareResult.value.smb_user;
		}
		return '';
	});

	const getSharePassword = computed(() => {
		if (shareResult.value && shareResult.value.smb_password) {
			return shareResult.value.smb_password;
		}
		return '';
	});

	const createSMBShare = async () => {
		if (shareId.value) {
			try {
				await share.updateSMBShareMember(
					shareId.value,
					selectedUsers.value.map((e) => {
						return {
							id: e.id,
							permission: e.editingPermission || e.permission
						};
					}),
					smbPublic.value
				);
				notifySuccess(i18n.global.t('success'));
				store.closeHovers();
				const currentPath = filesStore.currentPath[origin_id];
				await filesStore.refushCurrentRouter(
					currentPath.path + currentPath.param,
					filesStore.activeMenu(origin_id).driveType,
					origin_id
				);
			} catch (error) {
				console.log(error);
			}

			return;
		}
		try {
			const index = filesStore.selected[origin_id][0];
			const file = filesStore.getTargetFileItem(index, origin_id);

			if (!file) {
				return false;
			}

			const option = {
				name: decodeURI(file.name),
				share_type: ShareType.SMB,
				permission: smbPublic.value
					? SharePermission.Edit
					: readOnly.value
					? SharePermission.View
					: SharePermission.Edit,
				password: '',
				expire_in: 0,
				expire_time: '',
				users: smbPublic.value
					? undefined
					: selectedUsers.value.map((e) => {
							return {
								id: e.id,
								permission:
									e.editingPermission || e.permission || SharePermission.View
							};
					  }),
				public_smb: smbPublic.value
			};

			shareResult.value = await share.create(file, option);
		} catch (error) {
			return undefined;
		}
	};

	const createSMBUser = async () => {
		const result = await share.createSMBUser(
			smbAccount.value,
			smbPassword.value
		);
		if (result) {
			smbPassword.value = '';
			smbAccount.value = '';
			initSMBUsers();
		}
		return result;
	};

	const onCancel = () => {
		store.closeHovers();
	};

	const initSMBUsers = async () => {
		const totalUsers = await share.getSMBUsers();
		smbUsers.value = [];
		smbMobileUsers.value = [];
		totalUsers.forEach((e) => {
			const selectedItem = selectedUsers.value.find((item) => item.id == e.id);
			if (!selectedItem) {
				smbUsers.value.push({
					name: e.name,
					id: e.id,
					selected: false
				});
			}
			smbMobileUsers.value.push({
				name: e.name,
				id: e.id,
				selected: selectedItem != undefined,
				permission: selectedItem
					? selectedItem.editingPermission || selectedItem.permission
					: SharePermission.View
			});
		});
		totalSMBUsers.value = totalUsers;
		return totalUsers;
	};

	const deleteItem = (item: SMBPermissionUser) => {
		selectedUsers.value = selectedUsers.value.filter((e) => e.id != item.id);

		smbUsers.value.push({
			name: item.name,
			selected: false,
			id: item.id
		});
	};

	const initShareInfo = async () => {
		if (filesStore.selected[origin_id].length > 0) {
			const index = filesStore.selected[origin_id][0];
			const file = filesStore.getTargetFileItem(index, origin_id);
			let users:
				| {
						name: string;
						id: string;
						permission: SharePermission;
				  }[]
				| undefined = undefined;
			if (file && file.isShareItem && file.id) {
				shareId.value = file.id;
				if (file.users) {
					users = file.users;
				}
				if (file.public_smb != undefined) {
					smbPublic.value = file.public_smb;
				}
			} else if (file) {
				const shareItem:
					| {
							users: {
								name: string;
								id: string;
								permission: SharePermission;
							}[];
							id: string;
							public_smb: boolean;
					  }
					| undefined
					| null = await share.getShareByFile(file, ShareType.SMB);
				if (shareItem) {
					smbPublic.value = shareItem.public_smb;
					shareId.value = shareItem.id;
					users = shareItem.users;
				}
			}
			if (users) {
				users.forEach((user) => {
					const selectedItem = selectedUsers.value.find(
						(item) => item.id == user.id
					);
					const mobileItem = smbMobileUsers.value.find(
						(item) => item.id == user.id
					);

					if (selectedItem) {
						selectedItem.permission = user.permission;
						selectedItem.editingPermission = user.permission;
					} else {
						selectedUsers.value.push({
							...user,
							editingPermission: user.permission,
							pwdDisplay: false
						});
						const selectedIndex = smbUsers.value.findIndex(
							(item) => item.id == user.id
						);
						if (selectedIndex >= 0) {
							smbUsers.value.splice(selectedIndex, 1);
						}
						// if (!selectedItem) {
						// 	smbUsers.value.push({
						// 		name: user.name,
						// 		id: user.id,
						// 		selected: false
						// 	});
						// }
					}
					if (mobileItem) {
						mobileItem.permission = user.permission;
					}
				});
			}
		}
	};

	return {
		smbPublic,
		readOnly,
		copy,
		getShareLink,
		getShareAccount,
		getSharePassword,
		createSMBShare,
		smbPassword,
		shareResult,
		onCancel,
		initSMBUsers,
		smbUsers,
		selectedUsers,
		totalSMBUsers,
		smbAccount,
		createSMBUser,
		deleteItem,
		currentStep,
		smbMobileUsers,
		initShareInfo
	};
}
