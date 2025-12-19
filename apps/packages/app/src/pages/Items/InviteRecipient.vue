<template>
	<div v-if="invite" class="inviteItemView bg-background-1">
		<div class="row justify-between header items-center">
			<q-icon
				v-if="isMobile"
				name="sym_r_chevron_left"
				size="24px"
				@click="goBack"
			/>
			<div v-else class="text-subtitle1 title">{{ t('invite') }}</div>

			<div class="tag text-body3 q-px-sm q-py-xs" :class="status?.class">
				<q-icon name="schedule" size="16px" />
				{{ status?.text }}
			</div>
		</div>

		<div class="stretch" :style="{ width: `${isMobile ? '100%' : '400px'}` }">
			<div class="text-subtitle2 title1">
				{{
					invite.purpose === 'confirm_membership'
						? t('vault_t.invitor_requested_to_confirm_your_membership_with', {
								inviter: invitor
						  })
						: t('vault_t.invitor_invited_to_join', {
								inviter: invitor
						  })
				}}
			</div>

			<div
				class="text-subtitle1 title2 row items-center justify-center q-my-md"
			>
				<q-icon name="sym_r_groups" size="40px" class="q-mr-sm" />
				{{ invite.org.name }}
			</div>

			<template v-if="invite.accepted || invite.expired">
				<div class="title4 text-subtitle3">
					{{
						invite.accepted
							? t('vault_t.you_have_already_accepted_this_invite')
							: t('this_invite_has_expired')
					}}
				</div>
			</template>

			<template v-else>
				<div class="title3 text-subtitle3">
					{{
						t(
							'vault_t.please_enter_the_confirmation_code_provided_to_you_by_the_organization_owner'
						)
					}}
				</div>

				<div class="codetip text-body3 q-mt-md text-grey-7">
					{{ t('confirmation_code') }}
				</div>
				<q-input
					outlined
					dense
					color="grey-6"
					v-model="code"
					class="codeInput"
				/>

				<q-btn
					class="confirm text-grey-9 q-mt-md"
					color="yellow-5"
					no-caps
					unelevated
					@click="onConfirm()"
					icon="sym_r_send"
				>
					{{ t('submit') }}
				</q-btn>
			</template>
		</div>
	</div>
</template>

<script lang="ts">
import {
	defineComponent,
	computed,
	ref,
	watch,
	onMounted,
	onUnmounted
} from 'vue';
import { Invite } from '@didvault/sdk/src/core';
import { formatDateTime } from '@didvault/sdk/src/util';
import { formatDateFromNow } from 'src/utils/format';
import { app } from '../../globals';
import { useQuasar } from 'quasar';
import { useRouter, useRoute } from 'vue-router';
import { useMenuStore } from '../../stores/menu';
import { notifyFailed, notifySuccess } from '../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { _t } from '../../utils/i18n';
import { useUserStore } from '../../stores/user';

export default defineComponent({
	name: 'InviteRecipient',
	props: {},
	components: {},
	setup() {
		const $q = useQuasar();
		const Route = useRoute();
		const router = useRouter();
		const meunStore = useMenuStore();
		const userStore = useUserStore();

		const inviteID = ref(Route.params.invite_id as string);
		const orgID = ref(Route.params.org_id as string);

		const code = ref<string>('');
		const invite = ref<Invite | null>(null);
		const isMobile = ref(
			process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile
		);

		const { t } = useI18n();

		async function updateInvite() {
			if (!inviteID.value || !orgID.value) {
				invite.value = null;
			} else {
				invite.value = await app.getInvite(orgID.value, inviteID.value);
			}
		}

		const invitor = computed(function () {
			return invite.value?.invitedBy
				? invite.value?.invitedBy.name
					? `${invite.value?.invitedBy.name} <${
							invite.value?.invitedBy.did
					  }@${userStore.getCurrentDomain()}>`
					: invite.value?.invitedBy.did
				: 'Someone';
		});

		watch(
			() => Route.params.invite_id,
			async (newValue, oldValue) => {
				if (newValue == oldValue) {
					return;
				}
				inviteID.value = newValue as string;
				await updateInvite();
			},
			{
				immediate: true
			}
		);

		watch(
			() => Route.params.org_id,
			async (newValue, oldValue) => {
				if (newValue == oldValue) {
					return;
				}
				orgID.value = newValue as string;
				await updateInvite();
			}
		);

		onMounted(async () => {
			meunStore.rightDrawerOpen = false;
			await updateInvite();
		});

		onUnmounted(() => {
			meunStore.rightDrawerOpen = true;
		});

		const status = computed(function () {
			return invite.value
				? invite.value?.expired
					? {
							icon: 'time',
							class: 'warning',
							text: t('this_invite_has_expired')
					  }
					: invite.value?.accepted
					? { icon: 'check', class: '', text: t('accepted') }
					: {
							icon: 'time',
							class: '',
							text: _t('expires_date_0', {
								date_0: formatDateFromNow(invite.value!.expires)
							})
					  }
				: null;
		});

		async function onConfirm() {
			if (!invite.value || !code.value) {
				notifyFailed(t('errors.please_input_the_code'));
				return;
			}

			try {
				const success = await app.acceptInvite(
					invite.value!,
					code.value.toLowerCase()
				);

				if (success) {
					notifySuccess(
						t('vault_t.successfully_accepted_the_invite_notify_info')
					);
					router.push('/items/');
				} else {
					notifyFailed(t('vault_t.wrong_confirmation_code_please_try_again'));
				}
			} catch (e) {
				console.error(e);
				notifyFailed(
					e.message || t('something_went_wrong_please_try_again_later')
				);
			}
		}

		async function onDismiss() {
			$q.dialog({
				title: t('dismiss_invite'),
				message: t('vault_t.are_you_sure_you_want_to_dismiss_this_invite'),
				cancel: true,
				persistent: true
			}).onOk(() => {
				router.push('/items/');
			});
		}

		const goBack = () => {
			router.go(-1);
		};

		return {
			t,
			formatDateFromNow,
			formatDateTime,
			invite,
			onConfirm,
			onDismiss,
			status,
			code,
			invitor,
			isMobile,
			goBack
		};
	}
});
</script>

<style lang="scss" scoped>
.inviteItemView {
	width: 100%;
	height: 100vh;
	padding: 0 20px;
	display: flex;
	flex-direction: column;
	background-color: $background-1;

	.header {
		width: 100%;
		height: 60px;
		.title {
			height: 60px;
			line-height: 60px;
		}
		.tag {
			border: 1px solid $grey-8;
			border-radius: 4px;
		}
	}

	.stretch {
		border: 1px solid $separator;
		border-radius: 8px;
		padding: 20px;
		margin: 20px auto 0;
		display: flex;
		flex-direction: column;
		align-items: center;
		.title1 {
			text-align: center;
			color: $grey-8;
		}
		.title2 {
			text-align: center;
			color: $blue-4;
		}
		.title3 {
			text-align: center;
			color: $grey-8;
		}
		.title4 {
			width: 350px;
			text-align: center;
			color: $green;
			margin: 16px 0;
			border: 1px solid $green;
			border-radius: 6px;
			padding: 4px 0;
			background: rgba(41, 204, 95, 0.1);
		}
		.codetip {
			width: 100%;
			line-height: 30px;
			text-align: left;
		}
		.codeInput {
			width: 100%;
		}
		.confirm {
			width: 100%;
		}
	}
}
</style>
