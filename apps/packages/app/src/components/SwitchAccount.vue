<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('switch_accounts')"
		:skip="false"
		:ok="false"
		:cancel="false"
		size="medium"
	>
		<div>
			<terminus-account-item
				v-for="(user, index) in userStore.users?.items"
				:key="'ii' + index"
				:user="user"
				@click="choose(user.id)"
				style="margin-top: 12px"
			>
				<template v-slot:side v-if="user.id === userStore.current_id">
					<q-icon
						name="sym_r_check_circle"
						size="24px"
						color="light-blue-default"
					/>
				</template>
			</terminus-account-item>

			<terminus-item
				img-bg-classes="bg-background-3"
				style="margin-top: 12px"
				icon-name="sym_r_person_add"
				:img-b-g-size="40"
				:border-radius="12"
				@click="addAccount"
				title-classes="add-new-terminus-name"
			>
				<template v-slot:title>
					{{ t('add_new_olares_id') }}
				</template>
			</terminus-item>
		</div>
	</bt-custom-dialog>
</template>
<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { app, clearSenderUrl, resetAPP, setSenderUrl } from './../globals';
import { useUserStore } from '../stores/user';
import { useI18n } from 'vue-i18n';
import TerminusAccountItem from './common/TerminusAccountItem.vue';
import TerminusItem from './common/TerminusItem.vue';
import { getAppPlatform } from '../application/platform';

const CustomRef = ref();

const router = useRouter();
const userStore = useUserStore();
const current_user = ref(userStore.current_user);
const { t } = useI18n();

const choose = async (id: string) => {
	if (id == current_user.value?.id) {
		CustomRef.value.onDialogCancel();
		return;
	}

	let user = userStore.users!.items.get(id)!;
	userStore.userUpdating = true;

	await app.lock(false);
	await userStore.setCurrentID(user.id);

	if (user.setup_finished) {
		setSenderUrl({
			url: user.vault_url
		});
	} else {
		clearSenderUrl();
	}

	resetAPP();

	await app.load(user.id);
	if (userStore.current_mnemonic?.mnemonic) {
		await app.unlock(userStore.current_mnemonic?.mnemonic);
	}
	userStore.userUpdating = false;
	CustomRef.value.onDialogCancel();

	router.replace('/connectLoading');
};

const addAccount = () => {
	if (getAppPlatform() && getAppPlatform().isPad) {
		router.push({
			path: '/setup/success'
		});
	} else {
		router.push({
			path: '/import_mnemonic'
		});
	}

	CustomRef.value.onDialogCancel();
};
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 400px;
		border-radius: 12px;

		.current-user {
			padding: 4px 8px;
			border-radius: 4px;
			text-align: center;
			border: 1px solid $blue-4;
			color: $blue-4;
		}
	}
}
</style>
