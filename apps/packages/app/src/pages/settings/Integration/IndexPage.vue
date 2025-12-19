<template>
	<page-title-component
		:show-back="false"
		:title="t(`home_menus.${MENU_TYPE.Integration.toLowerCase()}`)"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<app-menu-feature
			:menu-type="MENU_TYPE.Integration"
			:clickable="true"
			@click="onAddIntegrationAccount"
		/>
		<app-menu-feature
			image="settings/imgs/root/cookie.svg"
			:label="t('Manage Your Cookies')"
			:description="
				t('View and modify cookies to manage your site data in one place')
			"
			:clickable="true"
			@click="onCookieManagement"
		/>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AppMenuFeature from 'src/components/settings/AppMenuFeature.vue';
import { MENU_TYPE } from 'src/constant';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const router = useRouter();

const onCookieManagement = () => {
	router.push({
		path: '/integration/cookie'
	});
};

const onAddIntegrationAccount = () => {
	router.push({
		path: '/integration/accountList'
	});
};
</script>

<style scoped lang="scss">
.cookie-manage-icon {
	width: 40px;
	height: 40px;
	background: linear-gradient(180deg, #fcfcfc 0%, #f2f2f2 100%);
	border-radius: 50%;
}
.application-logo {
	width: 40px;
	height: 40px;
}

.application-name {
	color: $ink-1;
}

.application-label {
	color: $ink-2;
}

.add-btn {
	border-radius: 8px;
	padding: 6px 12px;
	border: 1px solid $separator;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn:hover {
	background-color: $background-3;
}
</style>
