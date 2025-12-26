<template>
	<terminus-title-bar :title="t('network')" />
	<TerminusScrollArea class="network-root">
		<template v-slot:content>
			<div class="q-mt-lg q-px-lg">
				<terminus-item icon-name="sym_r_public" :wholePictureSize="20">
					<template v-slot:title>
						<div class="text-subtitle2 text-ink-1">
							{{ t('Network status') }}
						</div>
					</template>
					<template v-slot:side>
						<terminus-user-status />
					</template>
				</terminus-item>
				<!-- <terminus-item
					icon-name="sym_r_avg_pace"
					:wholePictureSize="20"
					class="q-mt-md"
				>
					<template v-slot:title>
						<div class="text-subtitle2 text-ink-1">
							{{ t('Network Speed') }}
						</div>
					</template>
					<template v-slot:side>
						<div class="text-ink-1 text-body3 q-mr-md">
							<div class="row items-center justify-end">
								<span>--</span>
								<q-icon
									name="sym_r_arrow_upward_alt"
									color="positive"
									size="20px"
								/>
							</div>
							<div class="row items-center justify-end">
								<span>--</span>
								<q-icon
									name="sym_r_arrow_downward_alt"
									color="negative"
									size="20px"
								/>
							</div>
						</div>
					</template>
				</terminus-item> -->

				<terminus-item
					icon-name="sym_r_public_off"
					:wholePictureSize="20"
					class="q-mt-md"
				>
					<template v-slot:title>
						<div class="text-subtitle2 text-ink-1">
							{{ t('user_current_status.offline_mode.title') }}
						</div>
					</template>
					<template v-slot:side>
						<bt-switch
							size="sm"
							truthy-track-color="light-blue-default"
							v-model="offLineModeRef"
							@update:model-value="updateOffLineMode"
						/>
					</template>
				</terminus-item>

				<terminus-item
					icon-name="sym_r_shield_lock"
					:wholePictureSize="20"
					class="q-mt-md"
				>
					<template v-slot:title>
						<div class="text-subtitle2 text-ink-1">
							{{ t('user_current_status.vpn.title') }}
						</div>
					</template>
					<template v-slot:side>
						<bt-switch
							size="sm"
							truthy-track-color="light-blue-default"
							v-model="vpnToggleStatus"
						/>
					</template>
				</terminus-item>
			</div>
		</template>
	</TerminusScrollArea>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { useUserStore } from '../../../stores/user';

import { useI18n } from 'vue-i18n';
import TerminusItem from '../../../components/common/TerminusItem.vue';

import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';

import TerminusUserStatus from '../../../components/common/TerminusUserStatus.vue';
import { useScaleStore } from 'src/stores/scale';

const userStore = useUserStore();
const scaleStore = useScaleStore();
const { t } = useI18n();

onMounted(() => {});

const offLineModeRef = ref(userStore.current_user?.offline_mode || false);

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
.lock-slider {
	height: 60px;
	transition: height 0.5s;
	overflow: hidden;
	min-height: 0 !important;
	padding-top: 0px !important;
	padding-bottom: 0px !important;
}

.network-root {
	width: 100%;
	height: calc(100% - 56px);

	.lock-content {
		width: 100%;
		border: 1px solid $separator;
		background-color: $background-1;
		border-radius: 12px;

		&__header {
			height: 44px;

			&__title {
				margin-left: 16px;
			}
		}
	}
}
</style>
