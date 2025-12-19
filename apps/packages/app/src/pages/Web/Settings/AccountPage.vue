<template>
	<div class="row items-center">
		<q-avatar :icon="SETTING_MENU.account.icon" @click="toggleDrawer" />
		<q-toolbar-title>{{ t(SETTING_MENU.account.name) }}</q-toolbar-title>
	</div>

	<q-list class="q-ma-md accountCentent" separator>
		<q-item class="titleBg">
			<q-item-section>{{ t('profile') }}</q-item-section>
		</q-item>

		<q-item>
			<q-item-section>
				<q-input
					v-model="displayLocalName"
					borderless
					dense
					readonly
					label-color="blue-6"
					:label="t('email')"
					input-style="height: 44px;lineHeight: 44px"
				/>
			</q-item-section>
		</q-item>

		<q-item>
			<q-item-section>
				<q-input
					class="displayName"
					v-model="displayDomainName"
					debounce="100"
					borderless
					dense
					readonly
					label-color="blue-6"
					:label="t('prompts.displayName')"
					@update:model-value="updateInput"
					input-style="height: 44px;lineHeight: 44px"
				/>
			</q-item-section>
		</q-item>

		<q-item>
			<q-item-section>
				<q-item-label class="text-blue-6">{{
					t('mnemonic_phrase')
				}}</q-item-label>
				<div
					class="displayName cursor-pointer"
					style="margin-top: 10px"
					@click="copyMnemonic"
				>
					{{ mnemonic }}
				</div>
			</q-item-section>
		</q-item>

		<q-item
			class="accountfooter dispalyName row iterm-center justify-between"
			:class="{ show: changeDisplayNameFlag }"
		>
			<q-btn
				class="save web-confirm-btn but"
				:label="t('buttons.save')"
				@click="logOut"
			/>
			<q-btn
				class="cancel web-cancel-btn but"
				:label="t('buttons.cancel')"
				@click="deleteAccount"
			/>
		</q-item>
	</q-list>

	<div class="accountfooter q-ma-md row iterm-center justify-between">
		<q-btn
			class="text-color-title bg-grey-12 but"
			:label="t('logout')"
			@click="logOut"
		/>
		<q-btn
			class="web-bg-white text-ink-2 deleteAccount but"
			:label="t('delete_account')"
			type="reset"
			@click="deleteAccount"
		/>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref } from 'vue';
import { SETTING_MENU } from '../../../utils/constants';
import { useUserStore } from '../../../stores/user';
import { useMenuStore } from '../../../stores/menu';
import { getPlatform } from '@didvault/sdk/src/core';
import {
	notifyFailed,
	notifySuccess
} from '../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'AccountPage',
	components: {},
	setup() {
		const changeDisplayNameFlag = ref(false);
		const meunStore = useMenuStore();

		const userStore = useUserStore();

		const currentUser = userStore.users?.items.get(userStore.current_id || '');
		const currentMnemonic = userStore.users?.mnemonics.get(
			userStore.current_id || ''
		);

		const displayLocalName = ref(currentUser?.local_name || '');
		const displayDomainName = ref('@' + currentUser?.domain_name);
		const originDisplayName = ref('');
		const mnemonic = ref(currentMnemonic?.mnemonic);

		const updateInput = (data) => {
			if (data === originDisplayName.value) {
				changeDisplayNameFlag.value = false;
			} else {
				changeDisplayNameFlag.value = true;
			}
		};

		const logOut = () => {};

		const deleteAccount = () => {};

		const toggleDrawer = () => {
			meunStore.rightDrawerOpen = false;
		};

		const copyMnemonic = () => {
			getPlatform()
				.setClipboard(mnemonic.value)
				.then(() => {
					notifySuccess(t('copy_success'));
				})
				.catch(() => {
					notifyFailed(t('copy_fail'));
				});
		};

		const { t } = useI18n();

		return {
			SETTING_MENU,
			displayLocalName,
			displayDomainName,
			updateInput,
			logOut,
			deleteAccount,
			changeDisplayNameFlag,
			toggleDrawer,
			mnemonic,
			copyMnemonic,
			t
		};
	}
});
</script>

<style scoped lang="scss">
.accountCentent {
	border: 1px solid $separator;
	border-radius: 10px;
}

.accountfooter {
	.web-confirm-btn {
		background: $blue;
		color: $white;
	}

	.web-confirm-btn[disabled] {
		background: $grey-12;
		color: $grey-7;
	}

	.web-cancel-btn {
		background: $grey-11;
		color: $grey-7;
	}

	.but {
		width: 48%;
		height: 42px;
		border-radius: 10px;

		&:before {
			box-shadow: none;
		}
	}

	.web-bg-white {
		background: $white;
	}

	.deleteAccount {
		border: 1px solid $separator;
	}

	.save {
		width: 44%;
		margin-top: 8px;
		margin-left: 16px;
	}

	.cancel {
		width: 44%;
		margin-top: 8px;
		margin-right: 16px;
	}
}

.dispalyName {
	position: relative;
	top: 0;
	height: 0 !important;
	overflow: hidden;
	transition: all 0.5s;
	padding: 0;
	min-height: auto;
	padding: 0;
	border-top: 0 !important;
}

.show {
	height: 58px !important;
	border-top: 1px solid rgba(0, 0, 0, 0.12) !important;
}
</style>
