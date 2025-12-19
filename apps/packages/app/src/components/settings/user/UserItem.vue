<template>
	<div class="user-item-root column justify-start" v-if="account">
		<q-item
			clickable
			:class="deviceStore.isMobile ? 'user-item-mobile' : 'user-item'"
			class="item"
		>
			<q-item-section class="item-margin-left">
				<div class="row items-center">
					<q-avatar :size="`40px`">
						<TerminusAvatar
							:info="{
								terminusName: account?.terminusName,
								avatar: account?.avatar
							}"
							:size="40"
						/>
					</q-avatar>
					<div
						class="column justify-between"
						style="margin-left: 8px; height: 40px"
					>
						<div class="row justify-start">
							<div
								class="person-name"
								:class="
									deviceStore.isMobile ? 'text-subtitle2-m' : 'text-subtitle2'
								"
							>
								{{ account.name }}
							</div>
							<setting-connect-status
								v-if="!account.wizard_complete"
								:label="$t('not_activated')"
								icon="sym_r_devices_off"
							/>
						</div>

						<div class="text-body3 person-role">
							{{ account.roles ? getRoleName(account.roles.join('/')) : '' }}
						</div>
					</div>
				</div>
			</q-item-section>
			<q-item-section side class="item-margin-right">
				<q-icon name="sym_r_chevron_right" size="20px" />
			</q-item-section>
		</q-item>
		<bt-separator v-if="widthSeparator" />
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { AccountInfo } from 'src/constant/global';
import { getRoleName } from 'src/constant/index';
import SettingConnectStatus from 'src/components/settings/base/SettingConnectStatus.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import BtSeparator from '../base/BtSeparator.vue';

defineProps({
	account: {
		type: Object as PropType<AccountInfo>,
		require: true
	},
	widthSeparator: {
		type: Boolean,
		default: true
	},
	marginTop: {
		type: Boolean,
		default: true
	}
});
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.user-item-root {
	width: 100%;
	height: auto;

	.user-item {
		height: 64px;
		min-height: 64px;
		padding: 0;
	}

	.user-item-mobile {
		height: 64px;
		min-height: 64px;
		padding: 0;
	}

	.person-name {
		color: $ink-1;
	}

	.person-role {
		color: $ink-3;
	}
}
</style>
