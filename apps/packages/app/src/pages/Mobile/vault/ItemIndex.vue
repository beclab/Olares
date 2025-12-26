<template>
	<div class="vault-continer" v-if="userStore.isUnlocked">
		<VaultsDrawer />
		<div class="vault-content">
			<ItemList @toolabClick="onItemClicked" />
		</div>
	</div>
	<TermipassUnlockContent
		v-if="!userStore.isUnlocked"
		:cancel="false"
		:detailText="t('unlock.vault_unlock_introduce')"
		logo="login/vault_unlock.svg"
		:biometry-auto-unlock="true"
		class="bg-background-1"
	/>
	<router-view></router-view>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import VaultsDrawer from '../../../layouts/TermipassLayout/VaultsDrawer.vue';
import ItemList from './../../Items/ItemList.vue';
import TermipassUnlockContent from '../../../components/unlock/mobile/TermipassUnlockContent.vue';
import { useUserStore } from '../../../stores/user';
import { useI18n } from 'vue-i18n';
const Router = useRouter();

const userStore = useUserStore();
const { t } = useI18n();

function onItemClicked(itemid: string) {
	Router.push({
		path: '/items/' + (itemid ? itemid : '')
	});
}
</script>

<style lang="scss" scoped>
.vault-continer {
	width: 100%;
	height: 100%;
	.vault-content {
		width: 100%;
		height: 100%;
	}
}
</style>
