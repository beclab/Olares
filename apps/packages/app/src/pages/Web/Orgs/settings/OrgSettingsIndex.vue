<template>
	<div class="column itemlist">
		<div style="width: 100%; height: 60px">
			<div class="row items-center q-pl-md q-mt-sm">
				<q-icon :name="heading.icon" size="24px" />
				<div class="column q-pl-md">
					<div class="text-ink-1 text-caption">
						{{ org?.name }}
					</div>
					<div class="text-body1 text-weight-bold">
						{{ heading.title }}
					</div>
				</div>
			</div>
		</div>

		<div
			class="settingCard"
			:style="{ width: `${isMobile ? '90%' : '500px'}` }"
		>
			<div class="header text-subtitle1">{{ t('security') }}</div>
			<div class="btn" @click="rotateKeys()">
				{{ t('rotate_cryptographic_keys') }}
			</div>
		</div>

		<div
			class="settingCard"
			:style="{ width: `${isMobile ? '90%' : '500px'}` }"
		>
			<div class="header text-subtitle1">{{ t('buttons.more') }}</div>
			<div class="btn" @click="changeName()">
				{{ t('change_organization_name') }}
			</div>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, ref, onMounted, onUnmounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import { app } from '../../../../globals';
import { OrgRole, OrgMemberStatus } from '@didvault/sdk/src/core';
import { useMenuStore } from '../../../../stores/menu';
import {
	notifySuccess,
	notifyWarning
} from '../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

export default defineComponent({
	name: 'OrgSettingIndex',
	setup() {
		const $q = useQuasar();
		const meunStore = useMenuStore();
		const quota = ref();
		const isMobile = ref(
			process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile
		);
		const org = computed(function () {
			return app.orgs.find((org) => org.id == meunStore.org_id);
		});

		if (org.value) {
			quota.value = app.getOrgProvisioning(org.value).quota;
		}

		const isOwner = computed(function () {
			return app && app.account && org.value!.isOwner(app.account);
		});

		const heading = computed(function () {
			return {
				icon: 'sym_r_settings',
				title: 'Settings'
			};
		});

		const rotateKeys = () => {
			$q.dialog({
				title: t('confirm'),
				message: t('rotate_cryptographic_key_message'),
				cancel: true,
				persistent: true
			}).onOk(async () => {
				try {
					await app.rotateOrgKeys(org.value!);
					notifySuccess(t('rotate_cryptographic_key_notify_success'));
				} catch (e) {
					notifyWarning(t('something_went_wrong_please_try_again_later'));
				}
			});
		};

		const changeName = () => {
			$q.dialog({
				title: t('rename_organization'),
				message: '',
				prompt: {
					model: org.value!.name,
					type: 'text'
				},
				cancel: true,
				persistent: true
			}).onOk(async (name) => {
				if (!name) {
					throw t('please_enter_a_name');
				}

				await app.updateOrg(org.value!.id, async (org) => (org.name = name));
				return name;
			});
		};

		onMounted(() => {
			meunStore.rightDrawerOpen = false;
		});
		onUnmounted(() => {
			meunStore.rightDrawerOpen = true;
		});

		const { t } = useI18n();

		return {
			heading,
			org,
			isOwner,
			OrgRole,
			OrgMemberStatus,
			quota,
			rotateKeys,
			changeName,
			isMobile,
			t
		};
	}
});
</script>

<style lang="scss" scoped>
.itemlist {
	width: 100%;
	height: 100%;

	.settingCard {
		border: 1px solid $separator;
		border-radius: 6px;
		margin: 24px auto 0;
		.header {
			width: 100%;
			height: 38px;
			line-height: 38px;
			color: $ink-1;
			text-indent: 20px;
			border-bottom: 1px solid $separator;
			background-color: $tooltip-color;
		}
		.btn {
			width: 90%;
			height: 46px;
			line-height: 46px;
			text-align: center;
			color: $ink-1;
			background-color: $tooltip-color;
			margin: 0 auto;
			border-radius: 6px;
			margin-top: 8px;
			margin-bottom: 8px;
			cursor: pointer;
			font-weight: 600px;
			&:hover {
				background-color: $tooltip-background;
			}
		}
	}
}
</style>
