<template>
	<div v-if="invite" class="inviteItemView bg-background-1">
		<div class="row justify-between header items-center">
			<q-icon
				v-if="isMobile"
				name="sym_r_chevron_left"
				size="24px"
				@click="goBack"
			/>
			<div v-else class="text-subtitle2 title text-ink-1">
				{{ t('invite') }}
			</div>

			<div class="tag text-ink-2 text-overline" :class="status?.class">
				<q-icon :name="status?.icon" size="20px" />
				{{ status?.text }}
			</div>
		</div>

		<div class="stretch">
			<div class="title1 text-body3">
				{{
					invite?.purpose === 'confirm_membership'
						? t('a_membership_confirmation_request_was_send_to')
						: t('an_invite_was_send_to')
				}}
			</div>

			<div class="title2 text-subtitle2">{{ invite?.did }}</div>

			<div class="title3 text-subtitle3">
				{{
					t(
						'following_confirmation_code_shuould_communicate_to_them_separately'
					)
				}}
			</div>

			<div class="title4 text-h5">
				{{ secret }}
			</div>

			<div class="bottom">
				<q-btn
					v-if="invite?.accepted"
					:disabled="invite?.expired"
					outline
					no-caps
					class="full-width resend"
					@click="onConfirm()"
				>
					<span class="row justify-center text-ink-1">
						<q-icon name="sym_r_mail" class="q-mr-xs" />
						{{
							invite?.purpose === 'confirm_membership'
								? t('confirm')
								: t('add_member')
						}}
					</span>
				</q-btn>

				<q-btn
					v-else
					outline
					no-caps
					color="yellow-6"
					class="full-width"
					@click="onResend()"
				>
					<span class="row justify-center text-ink-1">
						<q-icon name="sym_r_mail" class="q-mr-xs" />
						{{ t('resend') }}
					</span>
				</q-btn>
			</div>
		</div>
	</div>

	<div
		class="inviteItemView text-ink-2 row items-center justify-center bg-background-1"
		v-else
	>
		<img src="../../../../assets/layout/nodata.svg" />
		<span class="q-mb-md text-ink-2" style="margin-top: 32px">
			{{ t('no_invite_selected') }}
		</span>
	</div>
</template>

<script lang="ts">
import {
	defineComponent,
	computed,
	ref,
	onMounted,
	watch,
	onUnmounted
} from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { UnlockedOrg, UnlockedAccount } from '@didvault/sdk/src/core';
import { formatDateTime } from '@didvault/sdk/src/util';
import { formatDateFromNow } from 'src/utils/format';
import { debounce } from '@didvault/sdk/src/core';
import { app } from '../../../../globals';
import { useQuasar } from 'quasar';
import { useMenuStore } from '../../../../stores/menu';
import { busOn, busOff } from '../../../../utils/bus';
import {
	notifyFailed,
	notifySuccess
} from '../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'OrgInvitesView',
	props: {},
	components: {},
	setup() {
		const { t } = useI18n();
		const route = useRoute();
		const router = useRouter();
		const $q = useQuasar();
		const meunStore = useMenuStore();
		const invite = ref();
		const isMobile = ref(
			process.env.PLATFORM == 'MOBILE' ||
				process.env.PLATFORM == 'BEX' ||
				$q.platform.is.mobile
		);

		const org = ref();

		const initOrg = () => {
			org.value = app.orgs.find((org) => org.id == meunStore.org_id);
		};

		const getInvite = () => {
			if (org.value && route.params.org_type) {
				return org.value.getInvite(route.params.org_type as string);
			} else {
				return null;
			}
		};

		const status = computed(function () {
			return invite.value
				? invite.value?.expired
					? {
							icon: 'sym_r_schedule',
							class: 'warning',
							text: t('this_invite_has_expired')
					  }
					: invite.value?.accepted
					? {
							icon: 'sym_r_check_small',
							class: '',
							text: t('accepted')
					  }
					: {
							icon: 'sym_r_schedule',
							class: '',
							text: t('expires_date_0', {
								date_0: formatDateFromNow(invite.value!.expires)
							})
					  }
				: null;
		});

		let secret = ref();

		async function updateSecert() {
			if (!invite.value || !org.value || !app.account || app.account.locked) {
				return '';
			}
			await org.value.unlock(app.account as UnlockedAccount);
			await invite.value.unlock((org.value as UnlockedOrg).invitesKey);
			secret.value = invite.value && invite.value.secret!;
		}

		watch(
			() => route.params.org_type,
			(newVaule: any, oldVaule: any) => {
				if (oldVaule == newVaule) {
					return;
				}
				invite.value =
					org.value && org.value.getInvite(route.params.org_type as string);

				if (newVaule) {
					updateSecert();
				}
			}
		);

		async function onResend() {
			try {
				const newInvite = (
					await app.createInvites(
						org.value!,
						[invite.value!.did],
						invite.value!.purpose
					)
				)[0];
				if (newInvite) {
					router.replace({
						path: '/org/Invites/' + (newInvite.id ? newInvite.id : '')
					});
					secret.value = newInvite.secret!;
				}
			} catch (e) {
				notifyFailed(e.message);
			}
		}

		async function onConfirm() {
			try {
				const member = await app.confirmInvite(invite.value!);
				notifySuccess(
					t('membe_was_successfully_added_to_your_organization', {
						member: member.did
					})
				);
				meunStore.org_invites = false;
				meunStore.org_members = true;
				route.params.org_type = member.id as any;
			} catch (e) {
				notifyFailed(e.message);
			}
		}

		async function stateUpdate() {
			await initOrg();
			invite.value = await getInvite();
		}

		const goBack = () => {
			router.go(-1);
		};

		onMounted(() => {
			stateUpdate();
			busOn('orgSubscribe', stateUpdate);
			invite.value =
				org.value && org.value.getInvite(route.params.org_type as string);
			updateSecert();

			meunStore.$subscribe(() => {
				updateItems();
			});
		});

		onUnmounted(() => {
			busOff('orgSubscribe', stateUpdate);
		});

		let updateItems = debounce(() => {
			invite.value = getInvite();
		}, 50);

		return {
			formatDateTime,
			onResend,
			onConfirm,
			org,
			invite,
			secret,
			status,
			isMobile,
			goBack,
			t
		};
	}
});
</script>

<style lang="scss" scoped>
.inviteItemView {
	width: 100%;
	height: 100%;
	padding: 0 20px;
	display: flex;
	flex-direction: column;

	.header {
		width: 100%;
		height: 60px;

		.title {
			height: 60px;
			line-height: 60px;
		}

		.tag {
			border: 1px solid $ink-2;
			padding: 0 8px;
			border-radius: 4px;
		}
	}

	.stretch {
		width: 100%;
		border: 1px solid $separator;
		border-radius: 6px;
		padding: 30px;
		margin-top: 20px;

		.title1 {
			text-align: center;
			color: $ink-2;
			font-weight: 600;
		}

		.title2 {
			text-align: center;
			color: $ink-1;
			font-weight: 800;
			margin: 10px 0;
		}

		.title3 {
			text-align: center;
			color: $ink-2;
			font-weight: 600;
		}

		.title4 {
			text-align: center;
			color: $ink-1;
			font-weight: 800;
			margin: 16px 0;
		}

		.bottom {
			display: flex;
			align-items: center;
			justify-content: space-around;

			.resend {
				border: 1px solid separator;
			}
		}
	}
}
</style>
