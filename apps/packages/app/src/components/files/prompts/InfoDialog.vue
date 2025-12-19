<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.attributes')"
		size="medium"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		@onSubmit="onSubmit"
		@onCancel="onCancel"
		@onHide="onCancel"
	>
		<div class="dialog-desc">
			<q-tabs
				v-model="tab"
				class="text-ink-3"
				active-color="light-blue-default"
				align="left"
				no-caps
				:breakpoint="0"
			>
				<q-tab
					class="q-px-none q-mr-lg"
					content-class="tab-class"
					name="general"
					:label="t('files.general')"
				/>
				<q-tab
					v-if="permissionInDriveType.includes(currentFile.driveType)"
					class="q-px-none"
					content-class="tab-class"
					name="permission"
					:label="t('files.permission')"
				/>
			</q-tabs>

			<q-skeleton style="height: 1px" color="grey-5" />

			<div class="permission row justify-between items-center q-my-lg">
				<div
					class="row justify-start items-center"
					:style="{
						maxWidth: tab === 'general' ? '100%' : 'calc(100% - 120px)'
					}"
				>
					<terminus-file-icon
						class="q-mr-md"
						:name="currentFile.name"
						:type="attrInfo.type.value"
						:is-dir="currentFile.isDir"
						:iconSize="40"
					/>

					<div class="text-ink-1 text-subtitle1 text-ellipsis">
						{{ currentFile.name }}
					</div>
				</div>
				<q-select
					style="max-width: 120px"
					v-if="tab === 'permission'"
					class="permission-select"
					dense
					options-dense
					map-options
					emit-value
					borderless
					v-model="permission.uid"
					:options="permissionOption"
					dropdown-icon="sym_r_keyboard_arrow_down"
					color="ink-3"
					popup-content-class="options_selected_Account"
					popup-content-style="padding: 10px;"
				>
					<template v-slot:option="{ itemProps, opt, selected, toggleOption }">
						<q-item
							dense
							v-bind="itemProps"
							style="border-radius: 4px; overflow: hidden"
						>
							<q-item-section class="text-ink-2 select-popup">
								<q-item-label
									:class="{ 'text-light-blue-default': selected }"
									>{{ opt.label }}</q-item-label
								>
							</q-item-section>
							<q-item-section side>
								<q-checkbox
									:model-value="selected"
									checked-icon="sym_r_check_circle"
									@update:model-value="toggleOption(opt)"
								/>
							</q-item-section>
						</q-item>
					</template>
				</q-select>
			</div>

			<q-tab-panels v-model="tab" ref="panelsRef">
				<q-tab-panel
					class="q-px-none q-pt-none q-my-none q-pb-none"
					name="general"
				>
					<q-resize-observer @resize="onResize" />
					<div class="info-module column items-center">
						<template v-for="(value, key) in attrInfo" :key="key">
							<div
								class="info-item row justify-start items-center"
								v-if="value.show"
							>
								<span class="title text-ink-3 text-body3">{{
									value.label
								}}</span>
								<span class="detail text-ink-1 text-body3">
									<span class="detail-content">
										{{ value.value }}
									</span>

									<q-spinner-ios
										v-if="value.label === 'MD5' && md5Loading"
										color="light-blue-default"
										size="20px"
									/>

									<q-btn
										v-if="key === 'md5' || key == 'linkAddress'"
										class="btn-size-xs btn-no-text q-ml-sm"
										dense
										flat
										icon="sym_r_content_copy"
										color="light-blue-default"
										@click="copy(value.value)"
									>
										<q-tooltip>{{ t('copy') }}</q-tooltip>
									</q-btn>

									<q-btn
										v-if="key === 'originalPath'"
										class="btn-size-xs btn-no-text q-ml-sm"
										dense
										flat
										icon="sym_r_folder"
										color="light-blue-default"
										@click="enterOriginPath(value.value)"
									>
										<q-tooltip>{{ t('files.Original path') }}</q-tooltip>
									</q-btn>
								</span>
							</div>
						</template>
					</div>
				</q-tab-panel>

				<q-tab-panel
					v-if="permissionInDriveType.includes(currentFile.driveType)"
					class="q-px-none q-pt-none q-my-none q-pb-none"
					name="permission"
				>
					<q-resize-observer @resize="onResize" />

					<div
						class="check-box row items-center justify-start"
						v-if="currentFile.isDir"
					>
						<img
							v-if="permission.recursive"
							:src="activeImage"
							@click="toggleSelect"
						/>
						<img
							v-else-if="!$q.dark.isActive && !permission.recursive"
							:src="normalImage"
							@click="toggleSelect"
						/>
						<img v-else :src="normalDarkImage" @click="toggleSelect" />

						<span
							class="q-ml-sm text-ink-2 text-subtitle2"
							@click="toggleSelect"
						>
							{{ t('files.recursive_lookup') }}
						</span>
					</div>
				</q-tab-panel>
			</q-tab-panels>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref, onMounted, reactive, nextTick, PropType } from 'vue';
import { format } from 'quasar';
import { useI18n } from 'vue-i18n';
import { formatFileModified } from '../../../utils/file';
import { common, dataAPIs } from './../../../api';
import {
	notifySuccess,
	notifyFailed
} from '../../../utils/notifyRedefinedUtil';
import { useDataStore } from '../../../stores/data';
import { useOperateinStore } from '../../../stores/operation';
import {
	useFilesStore,
	FilesIdType,
	FileResType,
	shareTypeStr,
	sharePermissionStr
} from '../../../stores/files';
import { DriveType } from '../../../utils/interface/files';

import TerminusFileIcon from '../../common/TerminusFileIcon.vue';
import { SharePermission, ShareType } from 'src/utils/interface/share';
import { useRouter } from 'vue-router';
import { getApplication } from 'src/application/base';
import { decodeUrl, encodeUrl } from 'src/utils/encode';
const activeImage = './img/checkbox/check_box_blue.svg';
const normalImage = './img/checkbox/uncheck_box_light.svg';
const normalDarkImage = './img/checkbox/uncheck_box_dark.svg';

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	fileRes: {
		type: Object as PropType<FileResType | undefined>,
		required: false
	}
});

const { t } = useI18n();
const { humanStorageSize } = format;

const store = useDataStore();
const filesStore = useFilesStore();
const operationStore = useOperateinStore();
const router = useRouter();

const CustomRef = ref();
const tab = ref('general');
const md5Loading = ref(true);
const currentFile = ref(
	props.fileRes
		? props.fileRes
		: filesStore.getTargetFileItem(
				filesStore.selected[props.origin_id][0],
				props.origin_id
		  )
);
console.log('currentFile ===>', currentFile.value);

const attrInfo = reactive({
	type: {
		label: t('files.style'),
		value: '',
		show: !currentFile.value?.isShareItem
	},
	shared: {
		label: t('files.Shared'),
		value: '',
		show: !!currentFile.value?.isShareItem
	},
	path: {
		label: t('files.path'),
		value: '',
		show: true
	},
	originalPath: {
		label: t('files.Original path'),
		value: '',
		show:
			!!currentFile.value?.isShareItem && currentFile.value.shared_by_me == true
	},
	size: {
		label: t('files.size'),
		value: '',
		show: !currentFile.value?.isShareItem
	},
	md5: {
		label: 'MD5',
		value: '',
		show: !currentFile.value?.isShareItem
	},
	update_time: {
		label: t('files.update_time'),
		value: '',
		show: !currentFile.value?.isShareItem
	},
	shareScope: {
		label: t('files.Share scope'),
		value: '',
		show: !!currentFile.value?.isShareItem
	},
	shareOwner: {
		label: t('files.Owner'),
		value: '',
		show: !!currentFile.value?.isShareItem
	},
	sharePermission: {
		label: t('files.permission'),
		value: '',
		show: !!currentFile.value?.isShareItem
	},
	shareExpirationDate: {
		label: t('files.Expiration date'),
		value: '',
		show:
			!!currentFile.value?.isShareItem &&
			currentFile.value.share_type == ShareType.PUBLIC
	},
	linkAddress: {
		label: t('files.Link Details'),
		value: '',
		show:
			!!currentFile.value?.isShareItem &&
			(currentFile.value.share_type == ShareType.SMB ||
				currentFile.value.share_type == ShareType.PUBLIC)
	},
	accountList: {
		label: t('accounts'),
		value: '',
		show:
			!!currentFile.value?.isShareItem &&
			currentFile.value.share_type == ShareType.SMB
	}
});

const permission = reactive({
	path: '',
	uid: '1000',
	recursive: false
});

const permissionOption = ref([
	{
		label: 'Root',
		value: 0
	},
	{
		label: 'User',
		value: 1000
	}
]);

const md5InDriveType = [
	DriveType.Drive,
	DriveType.External,
	DriveType.Cache,
	DriveType.Data
];

const permissionInDriveType = [
	DriveType.Drive,
	DriveType.Data,
	DriveType.Cache
];

const toggleSelect = () => {
	permission.recursive = !permission.recursive;
};

const getPath = (item) => {
	return dataAPIs(item.driveType).getAttrPath(item);
};

const getOriginPath = (item) => {
	return dataAPIs(item.driveType).getOriginalPath(item);
};

const getPermission = async () => {
	const res = await operationStore.getPermission(currentFile.value);
	permission.uid = res;
};

const getMd5 = async () => {
	md5Loading.value = true;
	try {
		const res = await operationStore.getMd5(currentFile.value);
		attrInfo.md5.value = res;
		md5Loading.value = false;
	} catch (error) {
		md5Loading.value = false;
	}
};

const copy = (copyTxt: string) => {
	getApplication()
		.copyToClipboard(copyTxt)
		.then(() => {
			notifySuccess(t('copy_success'));
		})
		.catch(() => {
			notifyFailed(t('copy_fail'));
		});
};

const enterOriginPath = (path: string) => {
	const driveType = common().formatUrltoDriveType(path);
	const splitUrl = path.split('?');
	let param = '';
	if (splitUrl.length > 1) {
		param = '?' + splitUrl[1];
	}

	filesStore.setFilePath(
		{
			path: encodeUrl(splitUrl[0]),
			isDir: true,
			driveType: driveType || DriveType.Drive,
			param
		},
		false,
		true,
		props.origin_id
	);
};

const maxHeight = ref(0);
const panelsRef = ref();

const onResize = (size) => {
	maxHeight.value = Math.max(maxHeight.value, size.height);
	nextTick(() => {
		panelsRef.value.$el.querySelectorAll('.q-tab-panel').forEach((panel) => {
			panel.style.height = `${maxHeight.value}px`;
		});
	});
};

const initAttrInfo = () => {
	attrInfo.type.value = currentFile.value.isDir
		? t('files.folders')
		: currentFile.value.type;
	attrInfo.path.value = getPath(currentFile.value);
	attrInfo.size.value = currentFile.value.isDir
		? '-'
		: humanStorageSize(currentFile.value.size);
	attrInfo.update_time.value = formatFileModified(currentFile.value.modified);

	if (currentFile.value && currentFile.value.isShareItem) {
		attrInfo.shared.value = !!currentFile.value.shared_by_me
			? t('files.By Me')
			: t('files.With Me');
		attrInfo.shareScope.value = shareTypeStr(
			currentFile.value.share_type || ''
		);
		attrInfo.shareOwner.value = currentFile.value.owner || '';
		// ShareType.INTERNAL
		attrInfo.sharePermission.value = sharePermissionStr(
			currentFile.value.shared_by_me
				? SharePermission.ADMIN
				: currentFile.value.permission
		);
		attrInfo.shareExpirationDate.value =
			currentFile.value.shared_by_me &&
			(currentFile.value.share_type == ShareType.INTERNAL ||
				currentFile.value.share_type == ShareType.SMB)
				? '--'
				: formatFileModified(currentFile.value.expire_time);

		attrInfo.linkAddress.value =
			currentFile.value.isShareItem &&
			currentFile.value.share_type == ShareType.SMB
				? currentFile.value.smb_link
				: currentFile.value.isShareItem &&
				  currentFile.value.share_type == ShareType.PUBLIC
				? filesStore.getShareLinkAddress(currentFile.value.id)
				: '--';

		attrInfo.accountList.value =
			currentFile.value.isShareItem &&
			currentFile.value.share_type == ShareType.SMB &&
			!currentFile.value.public_smb &&
			currentFile.value.users
				? currentFile.value.users.map((e) => e.name).join(',')
				: '--';

		attrInfo.originalPath.value = decodeUrl(getOriginPath(currentFile.value));
	}
};

onMounted(() => {
	if (
		currentFile.value &&
		permissionInDriveType.includes(currentFile.value.driveType)
	) {
		getPermission();
	}

	if (currentFile.value.isDir) {
		attrInfo.md5.show = false;
	} else {
		if (md5InDriveType.includes(currentFile.value.driveType)) {
			getMd5();
		} else {
			attrInfo.md5.show = false;
		}
	}

	initAttrInfo();
});

const onCancel = () => {
	store.closeHovers();
};

const onSubmit = async () => {
	if (permissionInDriveType.includes(currentFile.value.driveType)) {
		await operationStore.setPermission(
			currentFile.value,
			permission.uid,
			permission.recursive
		);
	}

	store.closeHovers();
	CustomRef.value.onDialogOK();
};
</script>

<style>
.tab-class {
	div {
		font-size: 16px;
		font-weight: 500;
	}
}

.tab-content-class {
	font-family: Roboto;
	font-weight: 500;
	font-size: 20px;
	line-height: 24px;
	letter-spacing: 0%;
	vertical-align: middle;
}
</style>

<style lang="scss" scoped>
::v-deep(.q-tab) {
	.q-focus-helper {
		background-color: transparent !important;
		opacity: 0 !important;
	}
}
::v-deep(.q-checkbox__inner) {
	width: 20px;
	min-width: 20px;
	height: 20px;
	.q-checkbox__icon-container {
		width: 20px;
		height: 20px;
		right: 0;
	}
	&:before {
		background-color: transparent !important;
	}
}
.text-ellipsis {
	flex: 1;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.info-module {
	.info-item {
		width: 100%;
		margin-top: 12px;
		&:first-of-type {
			margin-top: 0;
		}

		.title {
			text-align: left;
			color: $prompt-message;
			width: 100px;

			text-overflow: ellipsis;
			white-space: nowrap;
			overflow: hidden;
		}

		.detail {
			width: calc(100% - 100px);
			display: inline-block;
			text-align: left;
			color: $ink-1;
			display: flex;
			align-items: center;
			justify-content: flex-start;
			.detail-content {
				display: inline-block;
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
			}
		}
	}
}

.permission-select {
	padding: 0 8px;
	border-radius: 8px;
	font-size: 16px;
	font-weight: 500;
	&:hover {
		background-color: $background-3;
	}
}

.select-popup {
	font-size: 16px;
	font-weight: 400;
}

.check-box {
	height: 32px;
	transition: width 0.5s ease;
	cursor: pointer;
	img {
		width: 20px;
		height: 20px;
	}
}
</style>
