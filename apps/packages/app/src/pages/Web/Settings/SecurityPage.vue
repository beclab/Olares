<template>
	<div class="row items-center">
		<q-avatar :icon="SETTING_MENU.security.icon" @click="toggleDrawer" />
		<q-toolbar-title> {{ t(SETTING_MENU.security.name) }} </q-toolbar-title>
	</div>

	<!-- <q-list class="q-ma-md sItem" separator>
		<q-item class="titleBg">
			<q-item-section>{{ t('master_password') }}</q-item-section>
		</q-item>
		<q-item>
			<q-item-section
				class="text-center cursor-pointer"
				@click="changePassword"
				>{{ t('Change Master Password') }}</q-item-section
			>
		</q-item>
	</q-list> -->

	<q-list class="q-ma-md sItem" separator>
		<q-item class="titleBg">
			<q-item-section>{{ t('auto_lock') }}</q-item-section>
		</q-item>
		<q-item class="row items-center justify-between">
			<q-item-label>{{ t('lock_automatically') }}</q-item-label>
			<q-item-label>
				<bt-switch
					size="sm"
					truthy-track-color="blue-default"
					v-model="lockStatus"
					@update:model-value="changeAutoLock"
				/>
			</q-item-label>
		</q-item>
		<q-item
			class="row items-center justify-between lock-slider"
			:class="!lockStatus ? 'hideSlider' : ''"
		>
			<q-item-label style="min-width: 30px">{{ t('after') }}</q-item-label>
			<q-slider
				v-model="lockTime"
				:step="1"
				:min="1"
				:max="10"
				style="margin: 0 20px 0 10px"
				color="blue-6"
				@change="changeAutoLockDelay"
			/>
			<q-item-label>{{ lockTime }}{{ t('min') }}</q-item-label>
		</q-item>
	</q-list>
</template>

<script lang="ts" setup>
import { ref, onMounted } from 'vue';
import { SETTING_MENU } from '../../../utils/constants';
import { useMenuStore } from '../../../stores/menu';
import { app } from '../../../globals';
import { useI18n } from 'vue-i18n';

const lockStatus = ref(app.settings.autoLock);

const selectionReport = ref([] as string[]);

const meunStore = useMenuStore();
const lockTime = ref(app.settings.autoLockDelay);

const toggleDrawer = () => {
	meunStore.rightDrawerOpen = false;
};

const changeAutoLockDelay = (value: any) => {
	app.setSettings({ autoLockDelay: value });
};

const changeAutoLock = (value: any) => {
	app.setSettings({ autoLock: value });
};

onMounted(() => {
	let securityReport = app.account?.settings.securityReport;
	for (const key in securityReport) {
		if (Object.prototype.hasOwnProperty.call(securityReport, key)) {
			const element = (securityReport as any)[key];
			if (element) {
				selectionReport.value.push(key);
			}
		}
	}
});

const { t } = useI18n();
</script>

<style lang="scss" scoped>
.sItem {
	border: 1px solid $separator;
	border-radius: 10px;
}
.biometric {
	.biometricItem {
		.section {
			flex: 1;
			padding-left: 8px;
			.sectionTag1 {
				padding: 3px 5px;
				border: 1px solid $blue-4;
				border-radius: 4px;
				margin-right: 10px;
			}
			.sectionTag2 {
				padding: 3px 5px;
				border: 1px solid $separator;
				border-radius: 4px;
			}
		}
	}
}

.lock-slider {
	height: 60px;
	transition: height 0.5s;
	overflow: hidden;
	min-height: 0 !important;
	padding-top: 0px !important;
	padding-bottom: 0px !important;
}

.hideSlider {
	height: 0 !important;
}
</style>
