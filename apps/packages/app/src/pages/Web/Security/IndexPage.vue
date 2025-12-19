<template>
	<q-page class="column report bg-white">
		<div
			class="row items-center q-pl-sm q-mt-md text-subtitle2 title"
			style="margin-bottom: 20px"
		>
			<q-icon
				v-if="isMobile"
				name="sym_r_chevron_left"
				size="24px"
				@click="goBack"
			/>
			<q-icon
				v-else
				name="sym_r_health_and_safety"
				size="24px"
				@click="toggleDrawer"
			/>
			<span class="q-ml-sm"> {{ t('security_report') }} </span>
		</div>
		<div class="scroll-containe">
			<q-scroll-area
				style="height: 100%"
				:thumb-style="scrollBarStyle.thumbStyle"
			>
				<div class="containe row">
					<div class="col-xs-12 col-sm-6">
						<q-card class="reportCard">
							<q-card-section
								class="title text-subtitle2 bg-grey-11 row items-center justify-between"
							>
								<span>{{ t('weak_passwords') }}</span>
								<BtIcon class="q-mr-md" src="info">
									<q-tooltip max-width="200px">
										{{ t('weak_password_info') }}
									</q-tooltip>
								</BtIcon>
							</q-card-section>
							<template v-if="weekPassword.length > 0">
								<q-card
									clickable
									v-ripple
									flat
									class="q-pa-sm repo-files cursor-pointer"
									v-for="(item, index) in weekPassword"
									:key="index"
								>
									<ReportCard :auditedItems="item" />
								</q-card>
							</template>
							<div class="text-center q-mt-lg text-ink-2 q-pb-lg" v-else>
								{{ t('nothing_found') }}
							</div>
						</q-card>
					</div>
					<div class="col-xs-12 col-sm-6">
						<q-card class="reportCard">
							<q-card-section
								class="title text-subtitle2 bg-grey-11 row items-center justify-between"
							>
								<span>{{ t('reused_passwords') }}</span>
								<BtIcon class="q-mr-md" src="info">
									<q-tooltip max-width="200px">
										{{ t('reused_passwords_info') }}
									</q-tooltip>
								</BtIcon>
							</q-card-section>
							<template v-if="reusedPassword.length > 0">
								<q-card
									clickable
									v-ripple
									flat
									class="q-pa-sm repo-files cursor-pointer"
									v-for="(item, index) in reusedPassword"
									:key="index"
								>
									<ReportCard :auditedItems="item" />
								</q-card>
							</template>
							<div class="text-center q-mt-lg text-ink-2 q-pb-lg" v-else>
								{{ t('nothing_found') }}
							</div>
						</q-card>
					</div>
					<div class="col-xs-12 col-sm-6">
						<q-card class="reportCard">
							<q-card-section
								class="title text-subtitle2 bg-grey-11 row items-center justify-between"
							>
								<span
									>{{ t('compromised_passwords') }} Compromised Passwords</span
								>
								<BtIcon class="q-mr-md" src="info">
									<q-tooltip max-width="200px">
										{{ t('compromised_passwords_info') }}
									</q-tooltip>
								</BtIcon>
							</q-card-section>
							<template v-if="compromisedPassword.length > 0">
								<q-card
									clickable
									v-ripple
									flat
									class="q-pa-sm repo-files cursor-pointer"
									v-for="(item, index) in compromisedPassword"
									:key="index"
								>
									<ReportCard :auditedItems="item" />
								</q-card>
							</template>
							<div class="text-center q-mt-lg text-ink-2 q-pb-lg" v-else>
								{{ t('nothing_found') }}
							</div>
						</q-card>
					</div>
					<div class="col-xs-12 col-sm-6">
						<q-card class="reportCard">
							<q-card-section
								class="title text-subtitle2 bg-grey-11 row items-center justify-between"
							>
								<span>{{ t('expired_items') }}</span>
								<BtIcon class="q-mr-md" src="info">
									<q-tooltip max-width="200px">
										{{ t('expired_items_info') }}
									</q-tooltip>
								</BtIcon>
							</q-card-section>
							<template v-if="expiredItems.length > 0">
								<q-card
									clickable
									v-ripple
									flat
									class="q-pa-sm repo-files cursor-pointer"
									v-for="(item, index) in expiredItems"
									:key="index"
								>
									<ReportCard :auditedItems="item" />
								</q-card>
							</template>
							<div class="text-center q-mt-lg text-ink-2 q-pb-lg" v-else>
								{{ t('nothing_found') }}
							</div>
						</q-card>
					</div>
				</div>
			</q-scroll-area>
		</div>
	</q-page>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useMenuStore } from '../../../stores/menu';
import { scrollBarStyle } from '../../../utils/contact';
import ReportCard from './ReportCard.vue';
import { app } from '../../../globals';
import { Vault, VaultItem } from '@didvault/sdk/src/core';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'SecurityIndex',
	components: {
		ReportCard
	},
	setup() {
		let mode = ref(1);
		const Router = useRouter();
		const $q = useQuasar();
		const meunStore = useMenuStore();
		const auditedItems = ref(app.auditedItems);

		const weekPassword = ref([] as { item: VaultItem; vault: Vault }[]);
		const reusedPassword = ref([] as { item: VaultItem; vault: Vault }[]);
		const compromisedPassword = ref([] as { item: VaultItem; vault: Vault }[]);
		const expiredItems = ref([] as { item: VaultItem; vault: Vault }[]);
		const isMobile = ref(
			process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile
		);

		onMounted(() => {
			getData();
		});

		const getData = () => {
			auditedItems.value.map((item) => {
				if (item.item.auditResults && item.item.auditResults.length > 0) {
					item.item.auditResults.map((cell) => {
						switch (cell.type) {
							case 'weak_password':
								weekPassword.value.push(item);
								break;
							case 'compromised_password':
								compromisedPassword.value.push(item);
								break;
							case 'reused_password':
								reusedPassword.value.push(item);
								break;
							case 'expired_item':
								expiredItems.value.push(item);
								break;

							default:
								break;
						}
					});
				}
			});
		};

		const toggleDrawer = () => {
			meunStore.leftDrawerOpen = !meunStore.leftDrawerOpen;
		};

		const goBack = () => {
			Router.go(-1);
		};

		const { t } = useI18n();

		return {
			mode,
			toggleDrawer,
			auditedItems,
			weekPassword,
			reusedPassword,
			compromisedPassword,
			expiredItems,
			isMobile,
			goBack,
			scrollBarStyle,
			t
		};
	}
});
</script>

<style scoped lang="scss">
.report {
	.title {
		padding-top: 20px;
		line-height: 16px;
	}

	.scroll-containe {
		height: calc(100vh - 100px);
		overflow: hidden;
		.containe {
			.reportCard {
				box-shadow: none;
				border: 1px solid $separator;
				border-radius: 10px;
				padding-top: 0;
				padding-left: 0;
				margin: 16px;

				.repo-files {
					border-bottom: 1px solid $separator;
					&:hover {
						background: $grey-1;
					}
				}
			}
		}
	}
}
</style>
