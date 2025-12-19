<template>
	<q-page class="column setting bg-white" :style-fn="myTweak">
		<div class="row items-center q-ml-lg" style="margin-bottom: 20px">
			<BtIcon src="settings" />
			<q-toolbar-title>{{ t('setting') }}</q-toolbar-title>
		</div>
		<q-list>
			<template v-for="(cell, index) in setting" :key="index">
				<q-item
					class="settingItem bg-blue-1-hover"
					:class="{
						'itemActive bg-blue-1': SETTING_MENU[cell].mode == mode
					}"
					clickable
					@click="setMode(SETTING_MENU[cell].mode)"
				>
					<q-item-section avatar>
						<q-avatar :icon="SETTING_MENU[cell].icon" />
					</q-item-section>
					<q-item-section>
						<q-item-label>{{ t(SETTING_MENU[cell].name) }}</q-item-label>
					</q-item-section>
				</q-item>
			</template>
		</q-list>
	</q-page>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useRouter } from 'vue-router';
import { SETTING_MENU } from '../../../utils/constants';
import { useI18n } from 'vue-i18n';

const settingMenu = ['security'];

let mode = ref(2);
const setting = ref(settingMenu);
const Router = useRouter();

const setMode = (menu: number) => {
	mode.value = menu;
	Router.push({
		path: '/settings/' + (mode.value ? mode.value : '')
	});
};
const myTweak = () => {
	return { height: '100%' };
};

const { t } = useI18n();
</script>

<style scoped lang="scss">
.setting {
	padding-top: 20px;
	height: 100vh;

	.settingItem {
		height: 58px;
		line-height: 58px;
		border-bottom: 0.5px solid #ececec;
		box-sizing: border-box;

		&.itemActive {
			border-left: 2px solid $blue;
		}
	}
}
</style>
