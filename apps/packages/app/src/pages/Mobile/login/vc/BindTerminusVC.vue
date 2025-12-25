<template>
	<!-- <terminus-title-bar :title="t('settings.safety')" /> -->
	<div class="bind-terminus-vc-root">
		<terminus-title-bar :title="t('Advanced account creation')">
			<template v-slot:right>
				<div
					class="scan-icon row items-center justify-center"
					@click="enterAccounts"
				>
					<q-icon name="sym_r_account_circle" size="24px" color="grey-8" />
				</div>
			</template>
		</terminus-title-bar>
		<terminus-scroll-area class="terminus-content-scroll-area padding-content">
			<template v-slot:content>
				<div class="content-root">
					<div class="home-module-title q-mt-md">
						{{ t('Create Olares ID with VC') }}
					</div>

					<!-- <terminus-item
						class="q-mt-md"
						img-bg-classes="bg-background-3"
						icon-name="sym_r_person"
						:whole-picture-size="32"
						@click="onCustomer()"
					>
						<template v-slot:title>
							<div class="text-subtitle1">
								{{ t('bind_personal_vc') }}
							</div>
						</template>
						<template v-slot:side>
							<div class="row items-center justify-end">
								<q-icon
									name="sym_r_keyboard_arrow_right"
									size="20px"
									color="ink-3"
								/>
							</div>
						</template>
					</terminus-item> -->

					<terminus-item
						class="q-mt-md"
						img-bg-classes="bg-background-3"
						icon-name="sym_r_groups"
						:whole-picture-size="32"
						@click="onOrg()"
					>
						<template v-slot:title>
							<div class="text-subtitle2">
								{{ t('bind_organization_vc') }}
							</div>
						</template>
						<template v-slot:side>
							<div class="row items-center justify-end">
								<q-icon
									name="sym_r_keyboard_arrow_right"
									size="20px"
									color="ink-3"
								/>
							</div>
						</template>
					</terminus-item>

					<div class="home-module-title q-mt-xl">
						{{ t('Set default domain') }}
					</div>

					<terminus-item
						class="q-mt-sm"
						v-for="domain in domains"
						:key="domain.value"
						@click="userStore.setDefaultDomain(domain.value)"
					>
						<template v-slot:title>
							<div class="text-subtitle2">
								{{ domain.name }}
							</div>
						</template>

						<template v-slot:side>
							<div class="q-mr-md">
								<q-img
									src="img/checkbox/check_box_circle.svg"
									width="14px"
									height="14px"
									v-if="domain.value === userStore.defaultDomain"
								/>
							</div>
						</template>
					</terminus-item>
				</div>
			</template>
		</terminus-scroll-area>
		<div class="bottom-view column padding-content">
			<TerminusExportMnemonicRoot :border="true" :height="48" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../../components/common/TerminusScrollArea.vue';
import TerminusItem from '../../../../components/common/TerminusItem.vue';
import { defaultDomains } from '../../../../utils/contact';
import { ref } from 'vue';
import { useUserStore } from '../../../../stores/user';
import TerminusExportMnemonicRoot from '../../../../components/common/TerminusExportMnemonicRoot.vue';

const { t } = useI18n();
const router = useRouter();

// const onCustomer = () => {
// 	router.push({ path: '/bind_customer_vc' });
// };

const onOrg = () => {
	router.push({ path: '/bind_org_vc' });
};

const domains = ref(defaultDomains);

const userStore = useUserStore();

const enterAccounts = () => {
	router.push('/accounts');
};
</script>

<style lang="scss" scoped>
.bind-terminus-vc-root {
	width: 100%;
	height: 100%;
	position: relative;

	.terminus-content-scroll-area {
		width: 100%;
		height: calc(100% - 56px - 48px - 48px);
	}

	.padding-content {
		padding-left: 20px;
		padding-right: 20px;
	}
}
</style>
