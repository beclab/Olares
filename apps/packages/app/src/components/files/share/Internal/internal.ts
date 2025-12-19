import * as filesUtil from 'src/api/files/v2/common/utils';
import { ShareItemUser, ShareMember, useFilesStore } from 'src/stores/files';
import share from 'src/api/files/v2/common/share';
import { ref } from 'vue';
import { SharePermission, ShareType } from 'src/utils/interface/share';

export enum ShareAddStep {
	BASE = 'base',
	INVITE_PEOPLE = 'invite_person',
	INVITE_PEOPLE_PERMISSION = 'invite_persion_permission',
	USER_PERMISSION = 'user_permission'
}

export function useInternalShare(origin_id: number) {
	const internalShareId = ref('');

	const filesStore = useFilesStore();

	const selectedUsers = ref([] as ShareItemUser[]);

	const CustomRef = ref();

	const cancelFocusDisable = ref(false);

	const currentStep = ref(ShareAddStep.BASE);

	const requestMember = async (file: {
		fileType?: string | undefined;
		fileExtend: string;
		oPath?: string | undefined;
	}) => {
		const shareItem:
			| {
					users: {
						name: string;
						permission: SharePermission;
					}[];
					id: string;
					owner: string;
			  }
			| undefined
			| null = await share.getShareByFile(file, ShareType.INTERNAL);
		if (shareItem) {
			return {
				owner: shareItem.owner,
				internalShareId: shareItem.id,
				members: shareItem.users.map((e) => {
					return {
						share_member: e.name,
						permission: e.permission,
						path_id: shareItem.id
					};
				})
			};
		}
	};

	const initUsers = async () => {
		await filesUtil.fetchUserList();

		let members = [] as ShareMember[];
		let owner: string | undefined = undefined;

		if (filesStore.shareRepoInfo || filesStore.selected[origin_id].length > 0) {
			if (filesStore.shareRepoInfo) {
				const data = await requestMember(filesStore.shareRepoInfo);
				if (data) {
					members = data.members;
					owner = data.owner;
					internalShareId.value = data.internalShareId;
				}
			} else {
				const index = filesStore.selected[origin_id][0];
				const file = filesStore.getTargetFileItem(index, origin_id);
				if (file && file.isShareItem && file.id) {
					internalShareId.value = file.id;
					members = await share.getMembers(file.id);
					owner = file.owner;
				} else if (file) {
					const data = await requestMember(file);
					if (data) {
						members = data.members;
						owner = data.owner;
						internalShareId.value = data.internalShareId;
					}
				}
			}
		}
		return {
			members,
			owner
		};
	};

	const getShareItem = () => {
		return (
			filesStore.shareRepoInfo ||
			filesStore.getTargetFileItem(filesStore.selected[origin_id][0], origin_id)
		);
	};

	const createInternalShare = async () => {
		try {
			const sharefile = getShareItem();
			if (!sharefile) {
				return;
			}

			const shareResult = await share.create(sharefile, {
				name: decodeURI(sharefile.name),
				share_type: ShareType.INTERNAL,
				permission: SharePermission.ADMIN,
				password: ''
			});
			internalShareId.value = shareResult.id;
			return shareResult.id;
		} catch (error) {
			return undefined;
		}
	};

	const createMembers = async (shareId: string) => {
		await share.addMember(
			shareId,
			selectedUsers.value
				.filter((e) => !e.isOwner)
				.map((e) => {
					return {
						share_member: e.name,
						permission:
							e.editingPermission || e.permission || SharePermission.View
					};
				})
		);
	};

	const updateShareMember = async () => {
		return await share.updateInternalShareMembers(
			internalShareId.value,
			selectedUsers.value.map((e) => {
				return {
					share_member: e.name,
					permission: e.editingPermission || e.permission
				};
			})
		);
	};

	return {
		internalShareId,
		initUsers,
		createInternalShare,
		selectedUsers,
		CustomRef,
		cancelFocusDisable,
		createMembers,
		currentStep,
		updateShareMember
	};
}
