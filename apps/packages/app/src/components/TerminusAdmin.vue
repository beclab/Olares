<template>
	<q-card class="user-info">
		<div class="header q-mt-md">
			<div class="fill"></div>
			<div class="users">
				<TerminusAvatar :info="userStore.terminusInfo()" :size="40" />
			</div>

			<div class="info q-mt-sm">
				<span class="text-h6 terminus-text-ellipsis local-name">{{
					current_user?.local_name
				}}</span>
				<span
					class="text-body3 terminus-text-ellipsis local-name text-ink-1 q-mt-xs"
					>{{ '@' + current_user?.domain_name }}</span
				>
			</div>

			<terminus-user-status class="q-mt-sm" />
		</div>

		<q-list class="q-px-xs">
			<q-item dense class="row items-center justify-between item-li">
				<div class="row items-center justify-between item-content">
					<div class="row items-center justify-start">
						<q-icon name="sym_r_sensors_off" size="20px" />
						<span class="text-body2 q-ml-xs">{{
							t('user_current_status.offline_mode.title')
						}}</span>
					</div>

					<bt-switch
						size="sm"
						truthy-track-color="blue-default"
						v-model="offLineModeRef"
						@update:model-value="updateOffLineMode"
					/>
				</div>
			</q-item>

			<q-item dense class="row items-center q-px-sm item-li">
				<div class="row items-center justify-between item-content">
					<div class="row items-center justify-start">
						<q-icon name="sym_r_sync_lock" size="20px" />
						<span class="text-body2 q-ml-xs">{{
							t('encrypted_connection')
						}}</span>
					</div>

					<bt-switch
						size="sm"
						truthy-track-color="blue-default"
						v-model="vpnToggleStatus"
					/>
				</div>
			</q-item>

			<q-item
				tag="label"
				v-ripple
				dense
				class="row items-center q-px-sm item-li"
				@click="changeAccount"
			>
				<div class="row items-center justify-between item-content">
					<div class="row items-center justify-start">
						<q-icon name="sym_r_swap_horizontal_circle" size="20px" />
						<span class="text-body2 q-ml-xs">{{ t('switch_accounts') }}</span>
					</div>
					<q-icon name="sym_r_chevron_right" size="20px" class="q-mr-sm" />
				</div>
			</q-item>

			<q-item
				tag="label"
				v-ripple
				dense
				class="row items-center q-px-sm item-li"
				@click="handleSettings"
			>
				<div class="row items-center justify-between item-content">
					<div class="row items-center justify-start">
						<q-icon name="sym_r_settings" size="20px" />
						<span class="text-body2 q-ml-xs">{{ t('settings.settings') }}</span>
					</div>
					<q-icon name="sym_r_chevron_right" size="20px" class="q-mr-sm" />
				</div>
			</q-item>
		</q-list>
	</q-card>
</template>
<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useUserStore } from '../stores/user';
import TerminusUserStatus from './common/TerminusUserStatus.vue';
import { useScaleStore } from '../stores/scale';
import { watch } from 'vue';
import { useI18n } from 'vue-i18n';

const emit = defineEmits(['switchAccount', 'handleSettings']);

const { t } = useI18n();

const userStore = useUserStore();
const scaleStore = useScaleStore();

const current_user = ref(userStore.current_user);
const offLineModeRef = ref(userStore.current_user?.offline_mode || false);

const changeAccount = () => {
	emit('switchAccount');
};

const handleSettings = () => {
	emit('handleSettings');
};

const updateOffLineMode = async () => {
	userStore.updateOfflineMode(offLineModeRef.value);
};

const vpnToggleStatus = computed({
	get: () => scaleStore.isOn || scaleStore.isConnecting,
	set: async (value) => {
		if (scaleStore.isDisconnecting) {
			return;
		}
		if (value) {
			await scaleStore.start();
		} else {
			await scaleStore.stop();
		}
	}
});
</script>

<style lang="scss" scoped>
.user-info {
	width: 320px;
	padding: 12px 8px;
	overflow: hidden;

	.header {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		position: relative;
		z-index: 0;

		.users {
			width: 40px;
			height: 40px;
			border-radius: 20px;
			overflow: hidden;
		}

		.info {
			display: flex;
			flex-direction: column;
			align-items: center;
			justify-content: center;
			width: calc(100% - 40px);

			.local-name {
				max-width: 100%;
			}
		}

		.fill {
			position: absolute;
			top: 0;
			left: 0;
			right: 0;
			margin: auto;
			width: 78px;
			height: 78px;
			z-index: -1;
			background: rgba(133, 211, 255, 0.7);
			filter: blur(50px);
		}
	}
	.item-li {
		border-radius: 8px;
		height: 36px;
		color: $ink-2;
		padding: 0 0 0 8px;

		.item-content {
			width: 100%;
		}
	}
}
</style>
