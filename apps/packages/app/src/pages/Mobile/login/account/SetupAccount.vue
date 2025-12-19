<template>
	<div class="setup-account-page">
		<terminus-title-bar v-if="userStore.current_user" />
		<terminus-scroll-area
			class="setup-account-page"
			:style="
				userStore.current_user ? 'height: calc(100% - 56px);' : 'height: 100%;'
			"
		>
			<template v-slot:content>
				<div
					class="terminus-login-page column justify-start items-center"
					style="position: relative"
				>
					<q-img
						class="setup-account-page__brand"
						:src="
							getRequireImage(
								$q.dark.isActive
									? 'login/larepass_brand_dark.svg'
									: 'login/larepass_brand.svg'
							)
						"
					/>
					<span class="setup-account-page__desc login-sub-title">{{
						t('Let us own our data again!')
					}}</span>
					<account-operations
						v-if="!isBex"
						class="setup-account-page__create"
						image-name="add"
						:title="t('Create an account')"
						@click="createDid"
					/>
					<account-operations
						class="setup-account-page__import"
						image-name="place_item"
						:title="t('Import an account')"
						@click="importDid"
					/>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';

import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../../../utils/imageUtils';
import AccountOperations from '../../../../components/common/AccountOperations.vue';
import { useQuasar } from 'quasar';
import { createUser } from './CreatUserBusiness';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import { useUserStore } from '../../../../stores/user';
import { notifyFailed } from '../../../../utils/notifyRedefinedUtil';
import { ref } from 'vue';

const $router = useRouter();
const { t } = useI18n();
const $q = useQuasar();

const userStore = useUserStore();

const isBex = ref(process.env.IS_BEX);

const createDid = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	const { message } = await createUser();
	if (message) {
		notifyFailed(message);
	} else {
		$router.replace({ path: '/BindTerminusName' });
	}
};

const importDid = () => {
	$router.push({ path: '/import_mnemonic' });
};
</script>

<style scoped lang="scss">
.setup-account-page {
	width: 100%;
	height: 100%;

	&__brand {
		margin-top: 80px;
		width: 96px;
		// height: 120px;
	}

	&__desc {
		margin-top: 20px;
		width: 335px;
		text-align: center;
	}

	&__create {
		width: calc(100% - 40px);
		margin-top: 64px;
	}

	&__import {
		width: calc(100% - 40px);
		margin-top: 20px;
	}
}
</style>
