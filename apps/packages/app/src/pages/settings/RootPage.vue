<template>
	<div
		class="settings-title text-h4 text-ink-1 row justify-between items-center"
	>
		{{ t('Setting') }}
	</div>
	<bt-scroll-area
		class="nav-height-scroll-area-conf"
		v-if="deviceStore.isMobile"
	>
		<q-list dense class="mobile-items-list" style="margin-top: 20px">
			<q-item
				clickable
				class="person-item row justify-start items-center item-padding-zero"
				@click="router.push({ path: '/person' })"
			>
				<q-item-section avatar class="q-pr-none item-margin-left">
					<setting-avatar :size="56" />
				</q-item-section>
				<q-item-section>
					<div class="column justify-between" style="margin-left: 8px">
						<div class="text-h5-m test-ink-1">
							{{ adminStore.user.name }}
						</div>
						<div class="text-body3-m text-ink-2">
							{{ '@' + adminStore.olaresId.split('@')[1] }}
						</div>
					</div>
				</q-item-section>
				<q-item-section side class="item-margin-right">
					<q-icon name="sym_r_keyboard_arrow_right" color="ink-1" />
				</q-item-section>
			</q-item>
		</q-list>

		<q-list
			v-for="(list, index) in adminStore.menus"
			:key="index"
			dense
			class="mobile-items-list"
			style="margin-top: 20px"
		>
			<div v-for="(item, index) in list" :key="item.key">
				<q-item
					class="item-padding-zero"
					style="height: 48px"
					clickable
					@click="changeItemMenu(item)"
				>
					<q-item-section
						avatar
						class="q-pr-none item-margin-left"
						style="min-width: 32px"
					>
						<q-img :src="item.img" width="27px" noSpinner />
					</q-item-section>

					<q-item-section
						class="text-subtitle2-m text-ink-1"
						style="margin-left: 8px"
					>
						{{ item.label }}
					</q-item-section>

					<q-item-section side class="item-margin-right">
						<q-icon name="sym_r_keyboard_arrow_right" color="ink-1" />
					</q-item-section>
				</q-item>
				<bt-separator v-if="index + 1 < list.length" :offset="20" />
			</div>
		</q-list>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import SettingAvatar from 'src/components/settings/base/SettingAvatar.vue';
import { useAdminStore } from 'src/stores/settings/admin';
import BtSeparator from 'src/components/settings/base/BtSeparator.vue';
import { useDeviceStore } from 'src/stores/settings/device';
import { useI18n } from 'vue-i18n';

const router = useRouter();
const adminStore = useAdminStore();
const deviceStore = useDeviceStore();

if (deviceStore.isMobile) {
	router.replace('');
} else {
	router.replace('/person');
}

const { t } = useI18n();

const changeItemMenu = (data: any): void => {
	const type = data.key;
	router.push({ name: type });
};
</script>

<style scoped lang="scss">
.settings-title {
	width: 100%;
	height: 56px;
	padding-left: 16px;
	padding-right: 16px;
	width: 100%;
}
.person-item {
	height: 80px;
	background-color: $background-2;
}
</style>
