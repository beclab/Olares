<template>
	<div
		class="column itemlist bg-background-1"
		:class="isWeb ? 'borderRight' : ''"
		v-if="!deviceStore.isScaning"
	>
		<div
			class="row items-center justify-between"
			style="height: 56px; width: 100%"
		>
			<item-list-search
				v-if="filterShowing === ItemHeaderModel.SEARCH"
				@search="search"
				@closeSearch="closeSearch"
			/>

			<item-list-default
				v-if="filterShowing === ItemHeaderModel.DEFAULT"
				@toolabClick="toolabClick"
				@changeHeaderModel="changeHeaderModel"
			/>

			<terminus-select-header
				v-if="filterShowing === ItemHeaderModel.CHECKBOX && selectIds"
				:showMove="true"
				:selectIds="selectIds"
				@handle-close="handleClose"
				@handle-select-all="handleSelectAll"
				@handle-remove="handleRemove"
				@handle-move="handleMove"
			/>
		</div>

		<q-list class="" style="width: 100%; height: calc(100% - 60px)">
			<TerminusUserHeaderReminder v-if="!isBex" />
			<template v-if="ids.length > 0">
				<!-- <q-scroll-area
					style="height: 100%"
					:thumb-style="scrollBarStyle.thumbStyle"
				> -->
				<div
					:class="isMobile ? 'mobile-list-padding' : 'web-list-padding'"
					style="height: 100%"
				>
					<terminus-select-all
						ref="terminusSelect"
						:items="ids"
						@show-select-mode="showSelectMode"
						@item-on-unable-select="itemOnUnableSelect"
					>
						<template v-slot="{ file }">
							<item-card-auth
								v-if="
									itemsMap[file.id].item.type === 3 ||
									itemsMap[file.id].item.icon === 'authenticator'
								"
								:selectIds="selectIds"
								:vaultItem="itemsMap[file.id]"
								@selectItem="selectItem"
							/>
							<item-card
								v-else
								:selectIds="selectIds"
								:vaultItem="itemsMap[file.id]"
								@selectItem="selectItem"
							/>
						</template>
					</terminus-select-all>
				</div>

				<div
					style="padding-bottom: 60px; width: 100%; height: 1px"
					v-if="
						termipassStore &&
						termipassStore.totalStatus &&
						termipassStore.totalStatus.isError == 2
					"
				/>
				<!-- </q-scroll-area> -->
			</template>

			<div
				class="column text-ink-2 items-center justify-center"
				style="height: 100%"
				v-if="ids.length == 0"
			>
				<img src="../../assets/layout/nodata.svg" />
				<span class="q-mb-md text-ink-2" style="margin-top: 32px">
					{{ t('vault_t.this_vault_don_not_have_any_items_yet') }}
				</span>
				<div
					class="newVault cursor-pointer q-px-md q-py-sm row items-center justify-center text-ink-2"
					@click="onCreate"
					v-if="canCreateItems"
				>
					<q-icon class="q-mr-sm" name="add" />
					<span>
						{{ t('vault_t.new_vault_item') }}
					</span>
				</div>
			</div>
		</q-list>

		<add-files
			v-if="isMobile && menuStore.currentItem === VaultMenuItem.AUTHENTICATOR"
			@addFile="addFile"
		/>
	</div>
	<ScanComponent
		v-if="deviceStore.isScaning"
		ref="scanCt"
		@cancel="scanCancel"
		@scan-result="scanResult"
	/>
</template>

<script lang="ts" setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute } from 'vue-router';
import { BtDialog } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { app } from '../../globals';
import {
	CryptoTemplate,
	VaultItem,
	escapeRegex,
	ExchangeTemplate,
	cloneItemTemplates,
	FieldType,
	Vault
} from '@didvault/sdk/src/core';
import { ListItem, ItemHeaderModel } from '@didvault/sdk/src/types';
import { useMenuStore } from '../../stores/menu';
import { VaultMenuItem } from '../../utils/contact';
import { autofillById } from '../../utils/bexFront';
// import { busOn, busOff } from '../../utils/bus';
import { useTermipassStore } from '../../stores/termipass';
import { useVaultStore } from '../../stores/vault';
import { getAppPlatform } from '../../application/platform';
import {
	bexFrontBusOn,
	bexFrontBusOff
} from '../../platform/interface/bex/utils';
import { addItem } from '../../platform/addItem';
import {
	notifyFailed,
	notifySuccess,
	notifyWarning
} from '../../utils/notifyRedefinedUtil';
import { VaultType } from '@didvault/sdk/src/core';

import MoveItemsPC from './dialog/MoveItemsPC.vue';
import MoveItemsMobile from './dialog/MoveItemsMobile.vue';
import ItemListSearch from './ItemListSearch.vue';
import ItemListDefault from './ItemListDefault.vue';
import ItemCard from './ItemCard.vue';
import ItemCardAuth from './ItemCardAuth.vue';
import CreateItem from './dialog/CreateItem.vue';
import ExchangeViewAdd from './dialog/ExchangeViewAdd.vue';
import CryptoViewAdd from './dialog/CryptoViewAdd.vue';
import TerminusSelectHeader from './../../components/common/TerminusSelectHeader.vue';
import TerminusUserHeaderReminder from '../../components/common/TerminusUserHeaderReminder.vue';
import TerminusSelectAll from './../../components/common/TerminusSelectAll.vue';
import AddFiles from './../Mobile/file/AddFiles.vue';
import DirOperationDialog from './DirOperationDialog.vue';
import ScanComponent from '../../components/common/ScanComponent.vue';
import { useDeviceStore } from '../../stores/device';
import throttle from 'lodash.throttle';
import { decodeAuthenticatorMigrationUrl } from 'src/platform/addItem';

function filterByString(fs: string, rec: VaultItem) {
	if (!fs) {
		return true;
	}
	const content = [
		rec.name,
		...rec.tags,
		...rec.fields.map((f) => f.name),
		...rec.fields.map((f) => f.value)
	]
		.join(' ')
		.toLowerCase();
	return content.search(escapeRegex(fs.toLowerCase())) !== -1;
}

const emits = defineEmits(['toolabClick']);

const $q = useQuasar();
const Route = useRoute();
const menuStore = useMenuStore();
const vaultStore = useVaultStore();
const termipassStore = useTermipassStore();
const deviceStore = useDeviceStore();

const filterInput = ref('');
const filterShowing = ref(ItemHeaderModel.DEFAULT);

const { t } = useI18n();

const isBex = ref(process.env.IS_BEX);
const isWeb = ref(process.env.APPLICATION == 'VAULT');
const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

const itemsMap = ref<Record<string, ListItem>>({});
const ids = ref<any[]>([]);
const terminusSelect = ref();
const selectIds = ref<null | any[]>(null);

const showSelectMode = (value: any | null) => {
	selectIds.value = value;
	if (value) {
		filterShowing.value = ItemHeaderModel.CHECKBOX;
	} else {
		filterShowing.value = ItemHeaderModel.DEFAULT;
	}
};

const handleClose = () => {
	if (terminusSelect.value) {
		terminusSelect.value.handleClose();
	}
	selectIds.value = null;
	filterShowing.value = ItemHeaderModel.DEFAULT;
};

const handleSelectAll = () => {
	if (terminusSelect.value) {
		terminusSelect.value.toggleSelectAll();
	}
};

const handleRemove = () => {
	removeItem();
};

const handleMove = () => {
	moveItems();
};

const changeHeaderModel = (value: ItemHeaderModel) => {
	filterShowing.value = value;
	if (value === ItemHeaderModel.CHECKBOX) {
		if (terminusSelect.value) {
			terminusSelect.value.intoCheckedMode();
		} else {
			selectIds.value = [];
		}
	}
};

const toolabClick = (value: any) => {
	emits('toolabClick', value);
};

async function onCreate() {
	let option: any = null;

	if (menuStore.currentItem === 'vault' && menuStore.vaultId) {
		const vaul = app.getVault(menuStore.vaultId);
		option = {
			value: vaul?.id,
			label: `${vaul?.org?.name || ''}${vaul?.org?.name ? ' / ' : ''}${
				vaul?.name
			}`
		};
	} else {
		option = {
			value: app.mainVault?.id,
			label: `${app.mainVault?.org?.name || ''}${
				app.mainVault?.org?.name ? ' / ' : ''
			}${app.mainVault?.name}`
		};
	}

	$q.dialog({
		component: CreateItem,
		componentProps: {
			option: option
		}
	}).onOk(async ({ selectedTemplate, vault }) => {
		if (selectedTemplate.id == 'exchange') {
			$q.dialog({
				component: ExchangeViewAdd,
				componentProps: {}
			}).onOk(async (exchange: ExchangeTemplate) => {
				await createItem(
					exchange.name,
					exchange.icon,
					exchange.fields,
					[],
					vault
				);
			});
		} else if (selectedTemplate.id == 'crypto') {
			$q.dialog({
				component: CryptoViewAdd,
				componentProps: {}
			}).onOk(async (template: CryptoTemplate) => {
				await createItem(
					template.name!,
					template.icon,
					template.fields,
					template.tags,
					vault
				);
			});
		} else {
			menuStore.isEdit = true;

			const editing_item = await createItem(
				'',
				selectedTemplate.icon,
				selectedTemplate.fields,
				[],
				vault
			);

			if (editing_item) {
				const id = editing_item.id;
				emits('toolabClick', id);
			}
			vaultStore.editing_item = editing_item;
		}
	});
}

async function createItem(
	name: string,
	icon: string,
	fields: { name: string; value?: string; type: FieldType }[],
	tags: string[],
	vault: Vault | null | undefined
) {
	const editing_item = await addItem(name, icon, fields, tags, vault);

	return editing_item;
}

const insertMap = (id: string, obj: ListItem) => {
	itemsMap.value[id] = obj;
	if (
		itemsMap.value[id].item.type === VaultType.TerminusTotp ||
		!itemsMap.value[id].item.name
	) {
		ids.value.unshift({
			id,
			selectedEnable: (value: string) => {
				return (
					itemsMap.value[value].item.type !== VaultType.TerminusTotp &&
					itemsMap.value[value].item.type !== VaultType.OlaresSSHPassword &&
					itemsMap.value[value].item.type !== VaultType.VC &&
					isEditable(itemsMap.value[value].vault)
				);
			}
		});
	} else {
		ids.value.push({
			id,
			selectedEnable: (value: string) => {
				return (
					itemsMap.value[value].item.type !== VaultType.TerminusTotp &&
					itemsMap.value[value].item.type !== VaultType.OlaresSSHPassword &&
					itemsMap.value[value].item.type !== VaultType.VC &&
					isEditable(itemsMap.value[value].vault)
				);
			}
		});
	}
};

const isEditable = (vault: Vault) => {
	const enable = vault ? app.isEditable(vault) : true;
	return enable;
};

const itemOnUnableSelect = (data: any) => {
	if (
		itemsMap.value[data].item.type == VaultType.TerminusTotp ||
		itemsMap.value[data].item.type == VaultType.OlaresSSHPassword ||
		itemsMap.value[data].item.type == VaultType.VC
	) {
		return;
	}
	if (!isEditable(itemsMap.value[data].vault)) {
		notifyFailed(t('Readonly'));
	}
};

const getItems = throttle(() => {
	getItemsDebunce();
}, 100);

async function getItemsDebunce() {
	itemsMap.value = {};
	ids.value = [];
	let filterUrl = '';

	if (process.env.PLATFORM == 'BEX') {
		const tab = await (getAppPlatform() as any).getCurrentTab();
		filterUrl = tab.url;
	}

	const filter = filterInput.value;

	let items = app.state.vaults.flatMap((vault) =>
		[...vault.items].map((item) => ({ vault, item }))
	);

	console.log('items ===>', items);

	if (filterUrl) {
		items = app.getItemsForUrl(filterUrl);
	}

	if (filter) {
		items = items.filter((item) => filterByString(filter, item.item));
	}
	sortByIsUpdateDate(items);
	updateItems(items);
}

function sortByIsUpdateDate(
	items: {
		vault: Vault;
		item: VaultItem;
	}[]
) {
	return items.sort((a, b) => {
		return a.item.updated.getTime() - b.item.updated.getTime();
	});
}

const canCreateItems = computed(() => {
	if (!menuStore.vaultId) {
		return true;
	}
	const vault = app.getVault(menuStore.vaultId);
	if (!vault) {
		return false;
	}
	return app.isEditable(vault);
});

const updateItems = (
	items: {
		vault: Vault;
		item: VaultItem;
	}[]
) => {
	const recentThreshold = new Date(
		Date.now() - app.settings.recentLimit * 24 * 60 * 60 * 1000
	);

	for (const { item, vault } of items) {
		if (menuStore.vaultId && vault.id !== menuStore.vaultId) {
			continue;
		}
		// if (!item.name) continue;

		const baseObj = {
			vault,
			item,
			section: '',
			firstInSection: false,
			lastInSection: false
		};
		if (menuStore.vaultId && vault.id === menuStore.vaultId) {
			insertMap(item.id, baseObj);
			continue;
		}

		if (process.env.PLATFORM !== 'MOBILE') {
			const removeTypeArray = [VaultType.VC, VaultType.TerminusTotp];
			if (removeTypeArray.includes(item.type)) {
				continue;
			}
		}

		switch (menuStore.currentItem) {
			case VaultMenuItem.ALLVAULTS:
				insertMap(item.id, baseObj);
				break;

			case VaultMenuItem.AUTHENTICATOR:
				if (item.type === 3 || item.icon === 'authenticator') {
					insertMap(item.id, baseObj);
				}
				break;
			case VaultMenuItem.RECENTLYUSED:
				if (
					app.state.lastUsed.has(item.id) &&
					app.state.lastUsed.get(item.id)! > recentThreshold
				) {
					insertMap(item.id, baseObj);
				}
				break;
			case VaultMenuItem.FAVORITES:
				if (app.account?.favorites.has(item.id)) {
					insertMap(item.id, baseObj);
				}
				break;
			case VaultMenuItem.ATTACHMENTS:
				if (!!item.attachments.length) {
					insertMap(item.id, baseObj);
				}
				break;
			case VaultMenuItem.MyVault:
				if (app.account?.id === vault.owner) {
					insertMap(item.id, baseObj);
				}
				break;
			default:
				if (item.tags.includes(menuStore.currentItem)) {
					insertMap(item.id, baseObj);

					break;
				}
		}
	}
};

async function selectItem(item: ListItem) {
	if (
		item.item.type == VaultType.OlaresSSHPassword ||
		item.item.type == VaultType.VC
	) {
		if (item.item.type == VaultType.OlaresSSHPassword) {
			notifyWarning(t('vault_t.SSH login password cannot be edited'));
		} else {
			notifyWarning(t('vault_t.VC data cannot be edited'));
		}

		return;
	}
	vaultStore.editing_item = null;

	if (process.env.PLATFORM == 'BEX') {
		autofillById(item.item.id);
		return;
	}
	if (item) {
		emits('toolabClick', item.item.id);
	}
}

// watch(
// 	() => Route.params.itemid,
// 	async (_newVal, oldVal) => {
// 		const currentItem = itemsMap.value[oldVal as string];
// 		if (currentItem && !currentItem.item.name) {
// 			await app.deleteItems([currentItem.item]);
// 			getItems();
// 		}
// 	}
// );

watch(
	() => vaultStore.editing_item,
	(newVal) => {
		if (newVal) {
			getItems();
		}
	}
);

onMounted(async () => {
	await getItems();
	vaultStore.editing_item = null;

	// busOn('appSubscribe', getItems);
	menuStore.$subscribe(() => {
		getItems();
	});

	if (isBex.value) {
		bexFrontBusOn('VAULT_TAB_UPDATE', getItems);
	}
});
onUnmounted(() => {
	// busOff('appSubscribe', getItems);
	if (isBex.value) {
		bexFrontBusOff('VAULT_TAB_UPDATE', getItems);
	}
});

async function search(value: string) {
	filterInput.value = value || '';
	await getItems();
}

function closeSearch() {
	filterInput.value = '';
	getItems();
	filterShowing.value = ItemHeaderModel.DEFAULT;
}

const removeItem = async () => {
	BtDialog.show({
		title: t('delete_vault'),
		message: t('are_you_sure_to_delete'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		platform: process.env.PLATFORM == 'MOBILE' ? 'mobile' : 'web',
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (res) {
				const deleteItems: VaultItem[] = [];
				for (const vault of app.state.vaults) {
					for (const item of vault.items) {
						if (
							selectIds.value &&
							selectIds.value.find((cell) => cell === item.id) &&
							item.type !== 3
						) {
							deleteItems.push(item);
						}
					}
				}
				app.deleteItems(deleteItems);
				terminusSelect.value.handleRemove();
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const hasCheckBox = ref<any[]>([]);

const moveItems = async () => {
	if (!selectIds.value || selectIds.value.length <= 0) {
		return false;
	}
	hasCheckBox.value = [];
	let hasAttachments = false;

	for (let i = 0; i < selectIds.value.length; i++) {
		const element = selectIds.value[i];
		// const hasCheckItem = itemList.value.find((a) => a.item.id === element);
		const hasCheckItem = itemsMap.value[element];

		if (
			hasCheckItem &&
			hasCheckItem.item.attachments &&
			hasCheckItem.item.attachments.length > 0
		) {
			hasAttachments = true;
		} else {
			hasCheckBox.value.push(hasCheckItem);
		}
	}

	if (hasAttachments) {
		await checkAttachments();
	} else {
		await checkHasCheckBox();
	}
};

const checkAttachments = () => {
	BtDialog.show({
		title: t('confirm'),
		message: t('vault_t.some_items_not_move'),
		okStyle: {
			background: 'yellow-default',
			color: '#1F1F1F'
		},
		platform: process.env.PLATFORM == 'MOBILE' ? 'mobile' : 'web',
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel')
	})
		.then(async (res: any) => {
			if (!selectIds.value || selectIds.value.length <= 0) {
				return false;
			}
			if (res) {
				for (let i = 0; i < selectIds.value.length; i++) {
					const element = selectIds.value[i];

					const hasCheckItem = itemsMap.value[element];

					if (
						!hasCheckItem ||
						!hasCheckItem.item.attachments ||
						hasCheckItem.item.attachments.length <= 0
					) {
						hasCheckBox.value.push(hasCheckItem);
					}
				}
				await checkHasCheckBox();
			}
		})
		.catch((err: Error) => {
			console.log('click cancel', err);
		});
};

const checkHasCheckBox = () => {
	if (
		hasCheckBox.value &&
		hasCheckBox.value.some(({ vault }) => !app.hasWritePermissions(vault))
	) {
		BtDialog.show({
			title: t('confirm'),
			message: t('vault_t.some_items_not_have_white_access_message'),

			okStyle: {
				background: 'yellow-default',
				color: '#1F1F1F'
			},
			platform: process.env.PLATFORM == 'MOBILE' ? 'mobile' : 'web',
			cancel: true,
			okText: t('base.confirm'),
			cancelText: t('base.cancel')
		})
			.then(async (res: any) => {
				if (res) {
					hasCheckBox.value = hasCheckBox.value.filter(({ vault }) =>
						app.hasWritePermissions(vault)
					);
					showMoveItemsDialog();
				}
			})
			.catch((err: Error) => {
				console.log('click cancel', err);
			});
	} else {
		showMoveItemsDialog();
	}
};

const showMoveItemsDialog = () => {
	$q.dialog({
		component: process.env.PLATFORM == 'MOBILE' ? MoveItemsMobile : MoveItemsPC,
		componentProps: {
			selected: hasCheckBox.value,
			leftText: t('cancel'),
			rightText: t('vault_t.move_item')
		}
	})
		.onOk(() => {
			handleClose();
			resetCheckbox();
		})
		.onCancel(() => {
			handleClose();
			resetCheckbox();
		});
};

const resetCheckbox = () => {
	selectIds.value = [];
};

const addFile = () => {
	$q.dialog({
		component: DirOperationDialog,
		componentProps: {}
	}).onOk((value) => {
		if (value === 'createCode') {
			createCode();
		} else if (value === 'scanCode') {
			scanCode();
		}
	});
};

const createCode = async (secret?: string, name?: string, batch = false) => {
	const template = cloneItemTemplates().find(
		(template) => template.id === 'authenticator'
	);

	if (!template) {
		return;
	}

	if (name) {
		template.name = name;
	}

	if (secret) {
		template.fields[0].value = secret;
	}

	const editing_item: any = await addItem(
		name || '',
		template.icon,
		template.fields,
		[],
		app.mainVault,
		[],
		undefined
	);

	if ((editing_item.id && editing_item.name) || batch) {
		return;
	}

	menuStore.isEdit = true;

	const id = editing_item.id ? editing_item.id : editing_item.item.id;
	emits('toolabClick', id);
};

// const scanIng = ref(false);

const scanCode = () => {
	deviceStore.isScaning = true;
};

const scanCancel = () => {
	deviceStore.isScaning = false;
};

const scanResult = async (result: string) => {
	if (
		!result.startsWith('otpauth-migration:') &&
		!result.startsWith('otpauth')
	) {
		notifyFailed(t('errors.invalid_code_please_try_again'));
		return;
	}

	if (result.startsWith('otpauth-migration:')) {
		scanAuthenticatorResult(result);
	} else {
		addScanResultToVault(result, false);
		deviceStore.isScaning = false;
	}
};

const addScanResultToVault = (result: string, batch = false) => {
	let url = new URL(result);
	let params = new URLSearchParams(url.search);
	let secret = params.get('secret');
	const issuer = params.get('issuer');

	if (!secret) {
		return false;
	}

	let items = app.state.vaults.flatMap((vault) =>
		[...vault.items]
			.filter((e) => e.icon === 'authenticator')
			.map((item) => ({ vault, item }))
	);

	const namePath = (
		result.startsWith('otpauth://totp/')
			? result.substring(16)
			: result.substring(10)
	).split('?');

	const nameList = namePath[0].split(':');

	let name =
		(issuer ? issuer + ':' : '') +
		(nameList.length > 0 ? nameList[1] : nameList[0]);

	if (name) {
		try {
			name = decodeURIComponent(name);
		} catch (error) {
			/* empty */
		}
	}

	if (
		items.find((e) => e.item.name == name && e.item.fields[0].value == secret)
	) {
		return true;
	}

	createCode(secret, name);
	return true;
};

const scanCt = ref();

const scanAuthenticatorResult = async (result: string) => {
	try {
		const list = await decodeAuthenticatorMigrationUrl(result);

		list.forEach((item) => {
			addScanResultToVault(item, true);
		});
		notifySuccess(t('success'));
		setTimeout(() => {
			scanCt.value.checkScanPermissionAndStart();
		}, 1000);
	} catch (error) {
		console.log('error ===>', error);
	}
};
</script>

<style lang="scss" scoped>
.web-list-padding {
	padding: 0 12px;
}

.mobile-list-padding {
	padding: 0 20px;
}

.itemlist {
	width: 100%;
	height: 100%;
	&.borderRight {
		border-right: 1px solid $separator;
	}

	.searchWrap {
		width: 100%;
		height: 56px;
		line-height: 56px;
		text-align: center;

		.searchInput {
			padding: 0 8px;
			border: 1px solid $blue;
			border-radius: 10px;
			margin: 8px 16px;
			display: inline-block;
			display: flex;
			align-items: center;
			justify-content: center;
		}
	}

	.menuAcion {
		width: 32px;
		height: 32px;
		border-radius: 8px;
		background: rgba(246, 246, 246, 1);
	}

	.avator {
		width: 32px;
		height: 32px;
		border-radius: 16px;
		overflow: hidden;
	}

	.checkOperate {
		border-radius: 4px;
		padding: 4px;
	}

	.newVault {
		border-radius: 8px;
		border: 1px solid $yellow-default;

		&:hover {
			background: $background-hover;
		}
	}

	.authenticator {
		height: 94px;
		border-radius: 12px;
		margin: 12px 16px;
		padding: 20px;
		border: 1px solid #e0e0e0;
		background: linear-gradient(
			180deg,
			rgba(253, 255, 203, 0.3) 0%,
			rgba(236, 255, 135, 0.3) 100%
		);
	}
}
</style>
