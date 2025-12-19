<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Users.toLowerCase()}`)"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature :menu-type="MENU_TYPE.Users" />
		<bt-list v-show="accountStore.accounts.length > 0">
			<template
				v-for="(account, index) in accountStore.accounts"
				:key="account.uid"
			>
				<user-item
					:account="account"
					:margin-top="index !== 0"
					:width-separator="index !== accountStore.accounts.length - 1"
					@click="pushToUserInfo(account)"
				/>
			</template>
		</bt-list>
		<list-bottom-func-btn
			v-show="accountStore.accounts.length > 0"
			@funcClick="createUser"
			class="q-mt-md"
			:title="t('create_account')"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import ListBottomFuncBtn from 'src/components/settings/ListBottomFuncBtn.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import CreateUserDialog from './dialog/CreateUserDialog.vue';
import UserItem from 'src/components/settings/user/UserItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useUserStore } from 'src/stores/settings/user';
import { AccountInfo } from 'src/constant/global';
import { MENU_TYPE } from 'src/constant';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { onMounted } from 'vue';

const accountStore = useUserStore();
const quasar = useQuasar();
const $router = useRouter();
const { t } = useI18n();

const pushToUserInfo = (account: AccountInfo) => {
	$router.push(`user/info/${account.name}`);
};

const createUser = () => {
	quasar
		.dialog({
			component: CreateUserDialog
		})
		.onOk(() => {
			accountStore.get_accounts();
		});
};

onMounted(() => {
	accountStore.get_accounts();
});
</script>

<style scoped lang="scss"></style>
