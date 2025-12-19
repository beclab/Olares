<template>
	<div class="dash-itemlist">
		<div style="width: 100%; height: 60px">
			<div class="row items-center justify-between">
				<div class="row items-center q-pl-md q-mt-sm">
					<q-icon
						v-if="platform === 'MOBILE' || platform === 'BEX'"
						name="sym_r_chevron_left"
						size="24px"
						@click="goBack"
					/>
					<q-icon :name="heading.icon" size="24px" />

					<div
						class="column q-pl-md"
						v-if="platform !== 'MOBILE' && platform !== 'BEX'"
					>
						<div class="text-ink-1 text-caption text-left">
							{{ org?.name }}
						</div>
						<div class="text-body2 text-weight-bold text-left">
							{{ heading.title }}
						</div>
					</div>
				</div>

				<div
					class="column q-pl-md"
					v-if="platform === 'MOBILE' || platform === 'BEX'"
				>
					<div class="text-ink-1 text-caption text-center">
						{{ org?.name }}
					</div>
					<div class="text-body2 text-weight-bold text-center">
						{{ heading.title }}
					</div>
				</div>

				<div class="row items-center q-py-xs q-my-md">
					<q-icon name="sym_r_add" size="28px" class="q-mr-md cursor-pointer">
						<q-tooltip>{{ t('add') }}</q-tooltip>
						<q-menu class="popup-menu">
							<q-list dense padding>
								<q-item
									class="row items-center justify-start text-body3 popup-item"
									clickable
									v-close-popup
									@click="createVault"
								>
									{{ t('new_vault') }}
								</q-item>
							</q-list>
						</q-menu>
					</q-icon>
				</div>
			</div>
		</div>

		<q-scroll-area
			style="height: calc(100% - 60px); width: 100%"
			:thumb-style="scrollBarStyle.thumbStyle"
		>
			<q-list class="cardList row justify-start">
				<div class="col-xs-12 col-sm-6 q-pa-md q-pb-none">
					<div class="listRow">
						<div class="header q-px-md q-py-sm row justify-between">
							<div class="items-center">
								<span class="text-ink-2 text-li-title">
									<q-icon name="sym_r_share_reviews" size="18px" />
									{{ t('invites') }}
								</span>
								<span class="q-ml-md text-ink-2">{{
									org?.invites.length
								}}</span>
							</div>
						</div>

						<div class="body">
							<q-scroll-area
								style="height: 100%"
								:thumb-style="scrollBarStyle.thumbStyle"
							>
								<div
									class="text-ink-2 column items-center justify-center"
									style="height: 304px"
									v-if="org?.invites.length == 0"
								>
									<img
										class="q-mt-sm"
										src="../../../../assets/layout/nodata.svg"
									/>
									<span class="q-mb-md text-ink-1" style="margin-top: 32px">
										{{ t('no_data') }}
									</span>
								</div>
								<template
									v-else
									v-for="(invite, index) in org?.invites"
									:key="'member' + index"
								>
									<div
										class="stretch q-px-lg q-py-sm"
										@click="openInvite(invite)"
									>
										<OrgInviteItem :invite="invite" />
									</div>
									<q-separator v-if="index < org?.invites.length - 1" />
								</template>
							</q-scroll-area>
						</div>
					</div>
				</div>

				<div class="col-xs-12 col-sm-6 q-pa-md q-pb-none">
					<div class="listRow column justify-start">
						<div class="header q-px-md q-py-sm row justify-between">
							<div class="items-center">
								<span class="text-ink-2 text-li-title">
									<q-icon name="sym_r_groups" size="18px" />
									{{ t('members') }}
								</span>
								<span class="q-ml-md text-ink-2"
									>{{ _availableMembers.length }}
									{{
										quota?.members !== -1 ? ` / ${quota?.members}` : ''
									}}</span
								>
							</div>
						</div>

						<div class="body">
							<q-scroll-area
								style="height: 100%"
								:thumb-style="scrollBarStyle.thumbStyle"
							>
								<div
									class="text-ink-2 column items-center justify-center"
									style="height: 304px"
									v-if="_availableMembers.length == 0"
								>
									<img
										class="q-mt-sm"
										src="../../../../assets/layout/nodata.svg"
									/>
									<span class="q-mb-md text-ink-1" style="margin-top: 32px">
										{{ t('no_data') }}
									</span>
								</div>
								<template
									v-else
									v-for="(member, index) in _availableMembers"
									:key="'member' + index"
								>
									<div
										class="stretch q-px-lg q-py-sm row items-center justify-start"
										v-if="member.status === OrgMemberStatus.Active"
										@click="openMember(member)"
									>
										<div class="q-mr-md avator">
											<TerminusAvatar
												:info="userStore.getUserTerminusInfo(member.id || '')"
												:size="28"
											/>
										</div>
										<div style="flex: 1">
											<div class="did text-subtitle1">
												{{ member.did }}
											</div>
											<div class="q-mt-xs">
												<span class="text-ink-1">{{ member.name }}</span>
												<span
													class="tag owner text-subtitle3"
													v-if="member.role === OrgRole.Owner"
												>
													{{ t('owner') }}
												</span>
												<span
													class="tag admin text-subtitle3"
													v-if="member.role === OrgRole.Admin"
												>
													{{ t('admin') }}
												</span>
												<span
													class="tag provisioned text-subtitle3"
													v-if="(member.status as string) === OrgMemberStatus.Provisioned"
												>
													{{ t('provisioned') }}
												</span>
												<span
													class="tag suspended text-subtitle3"
													v-if="org?.isSuspended(member)"
												>
													{{ t('isSuspended') }}
												</span>
											</div>
										</div>
									</div>
									<q-separator v-if="index < _availableMembers.length - 1" />
								</template>
							</q-scroll-area>
						</div>
					</div>
				</div>

				<div class="col-xs-12 col-sm-6 q-pa-md q-pb-none">
					<div class="listRow column justify-start">
						<div class="header q-px-md q-py-sm row justify-between">
							<div class="items-center">
								<span class="text-ink-2 text-li-title">
									<q-icon name="sym_r_apps" size="18px"> </q-icon>
									{{ t('Vault') }}
								</span>
								<span class="q-ml-md text-ink-2"
									>{{ org?.vaults.length }}
									{{ quota?.vaults !== -1 ? ` / ${quota?.vaults}` : '' }}
								</span>
							</div>
							<div>
								<BtIcon
									src="add"
									:width="16"
									:height="16"
									@click="createVault()"
								>
									<q-tooltip>{{ t('add_vault') }}</q-tooltip>
								</BtIcon>
							</div>
						</div>

						<div class="body">
							<q-scroll-area
								style="height: 100%"
								:thumb-style="scrollBarStyle.thumbStyle"
							>
								<div
									class="text-ink-2 column items-center justify-center"
									style="height: 304px"
									v-if="org?.vaults.length == 0"
								>
									<img
										class="q-mt-sm"
										src="../../../../assets/layout/nodata.svg"
									/>
									<span class="q-mb-md text-ink-1" style="margin-top: 32px">
										{{ t('this_organization_does_not_have_any_vaults_yet') }}
									</span>
								</div>
								<template
									v-else
									v-for="(vault, index) in org?.vaults"
									:key="'vault' + index"
								>
									<div class="card-wrap full-width">
										<q-card
											clickable
											v-ripple
											@click="openVault(vault)"
											flat
											class="vaultsCard row items-center justify-start q-px-md q-py-sm"
										>
											<q-card-section
												class="row items-center justify-between q-pa-none"
											>
												<q-icon name="sym_r_deployed_code" size="24px" />
											</q-card-section>
											<q-card-section
												class="column items-start justify-start q-pa-none q-ml-sm"
											>
												<div>{{ vault.name }}</div>
												<div class="row items-center justify-start">
													<div
														class="groups row items-center justify-center q-mr-sm"
													>
														<q-icon
															name="sym_r_groups_2"
															size="20px"
															class="q-mr-xs"
														/>
														{{ getGroups(vault)?.length }}
													</div>
													<div
														class="members text-body3 row items-center justify-center"
													>
														<q-icon
															name="sym_r_person"
															size="20px"
															class="q-mr-xs"
														/>
														{{ getMembers(vault)?.length }}
													</div>
												</div>
											</q-card-section>
										</q-card>
									</div>
									<q-separator v-if="index < org?.vaults.length - 1" />
								</template>
							</q-scroll-area>
						</div>
					</div>
				</div>
				<div class="col-xs-12 col-sm-6 q-pa-md q-pb-none"></div>
			</q-list>
		</q-scroll-area>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { app } from '../../../../globals';
import { scrollBarStyle } from '../../../../utils/contact';
import { OrgMember, OrgRole, OrgMemberStatus } from '@didvault/sdk/src/core';
import { useMenuStore } from '../../../../stores/menu';
import OrgInviteItem from '../invites/OrgInviteItem.vue';
// import VaultsMenu from '../../../../layouts/TermipassLayout/VaultsMenu.vue';
import { useUserStore } from '../../../../stores/user';
import { useI18n } from 'vue-i18n';
import { busOn, busOff } from '../../../../utils/bus';

export default defineComponent({
	name: 'OrgMembersIndex',
	components: {
		OrgInviteItem
		// VaultsMenu
		// OrgDashboardItem
	},
	setup() {
		const meunStore = useMenuStore();
		const userStore = useUserStore();
		const router = useRouter();
		const values = ref('');
		const vaultItemRef = ref();
		const showArrow = ref(false);
		const arrowItemObj = ref({});
		const contentStyle = ref({
			height: 0
		});
		const checkBoxArr = ref([]);
		const filterInput = ref('');
		const filterShowing = ref('default');
		const quota = ref();

		const platform = ref(process.env.PLATFORM);
		const org = ref();

		const initOrg = () => {
			org.value = app.orgs.find((org) => org.id == meunStore.org_id);
			quota.value = app.getOrgProvisioning(org.value).quota;
		};

		const getGroups = (vault) => {
			return org.value?.getGroupsForVault(vault);
		};

		const getMembers = (vault) => {
			return org.value?.getMembersForVault(vault);
		};

		const isOwner = computed(function () {
			return app && app.account && org.value!.isOwner(app.account);
		});

		const _availableMembers = computed(function () {
			return (
				(org.value &&
					org.value.members.filter(
						(member) =>
							member.status === OrgMemberStatus.Active &&
							org.value!.invites.find((o) => o.did !== member.did)
					)) ||
				[]
			);
		});

		const heading = computed(function () {
			return {
				icon: 'sym_r_dashboard',
				title: t('dashboard')
			};
		});

		async function createVault() {
			router.push({ path: '/org/Vaults' });
		}

		function openMember(item: OrgMember) {
			router.push({
				path: '/org/Members/' + (item.id ? item.id : '')
			});
			meunStore.org_mode_id = item.did;
			meunStore.currentItem = 'members';
		}

		function openInvite(item: OrgMember) {
			router.push({
				path: '/org/Invites/'
			});
			meunStore.org_mode_id = item.id!;
		}

		function openVault(item: any) {
			router.push({
				path: '/org/Vaults/'
			});
			meunStore.org_mode_id = item.id;
		}

		const goBack = () => {
			router.go(-1);
		};

		//let unsubscribe : any;
		onMounted(() => {
			meunStore.rightDrawerOpen = false;
			busOn('orgSubscribe', initOrg);
		});
		onUnmounted(() => {
			meunStore.rightDrawerOpen = true;
			busOff('orgSubscribe', initOrg);
		});

		const { t } = useI18n();

		return {
			createVault,
			openMember,
			openInvite,
			openVault,
			heading,
			filterShowing,
			filterInput,
			org,
			values,
			showArrow,
			arrowItemObj,
			vaultItemRef,
			contentStyle,
			checkBoxArr,
			isOwner,
			_availableMembers,
			OrgRole,
			OrgMemberStatus,
			quota,
			getGroups,
			getMembers,
			goBack,
			scrollBarStyle,
			platform,
			userStore,
			t
		};
	}
});
</script>

<style lang="scss" scoped>
.dash-itemlist {
	width: 100%;
	height: 100%;

	.cardList {
		width: 100%;

		// display: grid;
		// grid-template-columns: 48% 48%;
		// grid-template-rows: 400px 400px;
		.listRow {
			// width: 90%;
			border: 1px solid $separator;
			border-radius: 8px;
			cursor: pointer;

			.header {
				background-color: $tooltip-color;
				border-bottom: 1px solid $grey-12;
			}

			.body {
				height: 304px;

				.stretch {
					width: 100%;

					.avator {
						width: 28px;
						height: 28px;
						border-radius: 14px;
						overflow: hidden;
					}

					.did {
						width: 90%;
						white-space: nowrap;
						overflow: hidden;
						text-overflow: ellipsis;
						color: $ink-1;
					}

					.tag {
						padding: 0 4px;
						border-radius: 4px;
						margin-left: 10px;

						&.owner {
							border: 1px solid $orange;
							color: $orange;
						}

						&.admin {
							border: 1px solid $blue-4;
							color: $blue-4;
						}

						&.provisioned {
							border: 1px solid $blue-grey-8;
							color: $blue-grey-8;
						}

						&.suspended {
							border: 1px solid $blue-grey-8;
							color: $blue-grey-8;
						}
					}

					&:hover {
						background-color: $blue-grey-1;
					}
				}

				.card-wrap {
					display: flex;
					align-items: center;
					justify-content: center;

					.vaultsCard {
						width: 100%;
						border: 0;
						border-radius: 0;
						box-sizing: border-box;
						position: relative;
						border-radius: 8px;
						cursor: pointer;

						&:hover {
							background: $grey-1;
						}

						.tag {
							border: 1px solid $grey-7;
							padding: 0 4px;
							border-radius: 4px;
							float: right;
							height: 26px;
							line-height: 26px;
						}

						.groups,
						.members {
							border: 1px solid $grey-12;
							border-radius: 4px;
							padding: 0px 6px;
						}
					}
				}
			}
		}
	}
}
</style>
