<template>
	<div
		v-if="isPad || !isMobile"
		class="row items-center q-pl-md"
		@click="toggleDrawer"
	>
		<div class="item-list-icon row items-center justify-center">
			<q-icon :name="heading.icon" size="20px" />
		</div>

		<div
			class="q-ml-xs"
			:class="
				$q.platform.is.mobile
					? 'mobile-title text-subtitle2 text-ink-1'
					: 'text-ink-1 text-subtitle2'
			"
		>
			{{ heading.title }}
		</div>
	</div>
	<div v-else class="row items-center" style="padding-left: 20px">
		<TerminusAccountAvatar />
		<div @click="toggleDrawer">
			<div class="text-ink-1 text-h6 user-header__title">
				{{ t('Vault') }}
			</div>
			<div class="text-ink-3 text-subtitle3 row items-center">
				{{ heading.title }}
				<q-icon name="chevron_right" size="16px" />
			</div>
		</div>
	</div>
	<div class="row items-center q-py-xs text-ink-1 q-pr-md">
		<q-btn
			v-if="!isBex"
			class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
			icon="sym_r_checklist"
			text-color="ink-2"
			:disable="!canCreateItems"
			@click="changeHeaderModel(ItemHeaderModel.CHECKBOX)"
		>
			<q-tooltip v-if="canCreateItems">{{ t('select') }}</q-tooltip>
		</q-btn>

		<q-btn
			class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
			icon="sym_r_add"
			text-color="ink-2"
			@click="onCreate"
			:disable="!canCreateItems"
		>
			<q-tooltip v-if="canCreateItems">{{ t('create') }}</q-tooltip>
		</q-btn>

		<q-btn
			class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
			icon="sym_r_search"
			text-color="ink-2"
			@click="changeHeaderModel(ItemHeaderModel.SEARCH)"
		>
			<q-tooltip>{{ t('search') }}</q-tooltip>
		</q-btn>

		<div v-if="isMobile" class="q-mr-md" />
	</div>
</template>

<script lang="ts" setup>
import { ref, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { ItemHeaderModel } from '@didvault/sdk/src/types';
import {
	CryptoTemplate,
	ExchangeTemplate,
	Field,
	VaultItem
} from '@didvault/sdk/src/core';
import { app } from '../../globals';
import { useMenuStore } from '../../stores/menu';
import { VaultMenuItem } from '../../utils/contact';
import { useVaultStore } from '../../stores/vault';
import { getAppPlatform } from '../../application/platform';

import CreateItem from './dialog/CreateItem.vue';
import ExchangeViewAdd from './dialog/ExchangeViewAdd.vue';
import CryptoViewAdd from './dialog/CryptoViewAdd.vue';
import TerminusAccountAvatar from '../../components/common/TerminusAccountAvatar.vue';
import { addItem } from '../../platform/addItem';

const emits = defineEmits(['toolabClick', 'changeHeaderModel']);

const $q = useQuasar();
const menuStore = useMenuStore();
const vaultStore = useVaultStore();

const isMobile = ref(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile);

const isPad = getAppPlatform() && getAppPlatform().isPad;

const { t } = useI18n();

const isBex = ref(process.env.IS_BEX);

const changeHeaderModel = (value: ItemHeaderModel) => {
	emits('changeHeaderModel', value);
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
		console.log('onCreate vault', vault);
		if (selectedTemplate.id == 'exchange') {
			$q.dialog({
				component: ExchangeViewAdd,
				componentProps: {}
			}).onOk(async (exchange: ExchangeTemplate) => {
				await addItem(exchange.name, exchange.icon, exchange.fields, [], vault);
			});
		} else if (selectedTemplate.id == 'crypto') {
			$q.dialog({
				component: CryptoViewAdd,
				componentProps: {}
			}).onOk(async (template: CryptoTemplate) => {
				await addItem(
					template.name!,
					template.icon,
					template.fields,
					template.tags,
					vault
				);
			});
		} else {
			const editing_item = await addItem(
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
			menuStore.isEdit = true;
			vaultStore.editing_item = editing_item;
		}
	});
}

const heading = computed(function () {
	interface messageProp {
		title: string;
		superTitle: string;
		icon: string;
	}
	let message: messageProp = {
		title: '',
		superTitle: '',
		icon: ''
	};

	switch (menuStore.currentItem) {
		case VaultMenuItem.ALLVAULTS:
			if (menuStore.vaultId) {
				message = {
					title: app.getVault(menuStore.vaultId)?.name || 'Vault',
					superTitle: '',
					icon: 'sym_r_deployed_code'
				};
			} else {
				message = {
					title: t('vault_t.all_vaults'),
					superTitle: '',
					icon: 'sym_r_apps'
				};
			}
			break;
		case VaultMenuItem.AUTHENTICATOR:
			message = {
				title: t('vault_t.authenticator'),
				superTitle: '',
				icon: 'sym_r_encrypted'
			};
			break;
		case VaultMenuItem.RECENTLYUSED:
			message = {
				title: t('vault_t.recently_used'),
				superTitle: '',
				icon: 'sym_r_schedule'
			};
			break;
		case VaultMenuItem.FAVORITES:
			message = {
				title: t('favorites'), //Favorites
				superTitle: '',
				icon: 'sym_r_star'
			};
			break;
		case VaultMenuItem.ATTACHMENTS:
			message = {
				title: t('attachments'),
				superTitle: '',
				icon: 'sym_r_lab_profile'
			};
			break;
		case VaultMenuItem.MyVault:
			message = {
				title: t('vault_t.my_vault'),
				superTitle: '',
				icon: 'sym_r_frame_person'
			};
			break;
		case VaultMenuItem.TAGS:
			message = {
				title: menuStore.currentItem,
				superTitle: '',
				icon: 'sym_r_more'
			};
			break;
		default:
			message = {
				title: menuStore.currentItem,
				superTitle: '',
				icon: 'sym_r_apps'
			};
	}
	return message;
});

const toggleDrawer = () => {
	if (process.env.PLATFORM === 'MOBILE' || process.env.PLATFORM === 'BEX') {
		menuStore.leftDrawerOpen = !menuStore.leftDrawerOpen;
	}
};

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
</script>

<style lang="scss" scoped>
.mobile-title {
	display: flex;
	align-items: center;
}

.item-list-icon {
	width: 32px;
	height: 32px;
	border-radius: 4px;
}
</style>
