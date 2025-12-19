<template>
	<div class="plugin-options-layout row justify-center bg-background-1">
		<q-layout view="hHh lpR fFf" style="width: 1024px">
			<q-drawer
				show-if-above
				:width="280"
				side="left"
				bordered
				:breakpoint="0"
				class="bg-background-1"
			>
				<div class="q-mx-lg">
					<div class="row items-center flex-gap-x-md q-my-lg">
						<q-img
							:src="LarePassIcon"
							:ratio="1"
							spinner-size="0px"
							width="32px"
						/>
						<q-img :src="LarePassTextIcon" spinner-size="0px" width="92px" />
					</div>
					<bt-menu
						active-class="my-active-link"
						:items="initOption"
						v-model="active"
						style="width: 239px; padding: 0"
						@select="selectHandler"
					/>
				</div>
			</q-drawer>
			<q-page-container class="relative-position">
				<bt-scroll-area class="bex-options-scroll-wrapper" ref="scrollRef">
					<div class="text-h3 q-px-xxl q-mt-xxl q-mb-md text-ink-1">
						{{ currentItem.label }}
					</div>
					<div class="q-px-xxl q-pb-xl q-pt-md">
						<router-view />
					</div>
				</bt-scroll-area>
			</q-page-container>
		</q-layout>
	</div>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import AccountHeader from 'src/pages/Plugin/containers/AccountHeader.vue';
import accountHighlightIcon from 'src/assets/plugin/option-account-highlight.svg';
import accountIcon from 'src/assets/plugin/option-account.svg';
import accountHighlightDarkIcon from 'src/assets/plugin/option-account-highlight-dark.svg';
import accountDarkIcon from 'src/assets/plugin/option-account-dark.svg';

import appearanceHighlightIcon from 'src/assets/plugin/option-appearance-highlight.svg';
import appearanceIcon from 'src/assets/plugin/option-appearance.svg';
import appearanceHighlightDarkIcon from 'src/assets/plugin/option-appearance-highlight-dark.svg';
import appearanceDarkIcon from 'src/assets/plugin/option-appearance-dark.svg';

import collectHighlightIcon from 'src/assets/plugin/option-collect-highlight.svg';
import collectIcon from 'src/assets/plugin/option-collect.svg';
import collectHighlightDarkIcon from 'src/assets/plugin/option-collect-highlight-dark.svg';
import collectDarkIcon from 'src/assets/plugin/option-collect-dark.svg';

import translateHighlightIcon from 'src/assets/plugin/option-translate-highlight.svg';
import translateIcon from 'src/assets/plugin/option-translate.svg';
import translateHighlightDarkIcon from 'src/assets/plugin/option-translate-highlight-dark.svg';
import translateDarkIcon from 'src/assets/plugin/option-translate-dark.svg';

import securityHighlightIcon from 'src/assets/plugin/option-security-highlight.svg';
import securityIcon from 'src/assets/plugin/option-security.svg';
import securityHighlightDarkIcon from 'src/assets/plugin/option-security-highlight-dark.svg';
import securityDarkIcon from 'src/assets/plugin/option-security-dark.svg';
import LarePassIcon from '../../src-bex/icons/LarePass.png';
import LarePassTextIconLight from '../../src-bex/icons/larepass-text.png';
import LarePassTextDarkIcon from '../../src-bex/icons/larepass-text-dark.png';

import { useQuasar } from 'quasar';

const { t } = useI18n();
const $q = useQuasar();

const options = computed(() => {
	const accountDefaultIcon = $q.dark.isActive ? accountIcon : accountDarkIcon;
	const accountDefaultHighlightIcon = $q.dark.isActive
		? accountHighlightIcon
		: accountHighlightDarkIcon;

	const appearanceDefaultIcon = $q.dark.isActive
		? appearanceIcon
		: appearanceDarkIcon;
	const appearanceDefaultHighlightIcon = $q.dark.isActive
		? appearanceHighlightIcon
		: appearanceHighlightDarkIcon;

	const collectDefaultIcon = $q.dark.isActive ? collectIcon : collectDarkIcon;
	const collectDefaultHighlightIcon = $q.dark.isActive
		? collectHighlightIcon
		: collectHighlightDarkIcon;

	const translateDefaultIcon = $q.dark.isActive
		? translateIcon
		: translateDarkIcon;
	const translateDefaultHighlightIcon = $q.dark.isActive
		? translateHighlightIcon
		: translateHighlightDarkIcon;

	const securityDefaultIcon = $q.dark.isActive
		? securityIcon
		: securityDarkIcon;
	const securityDefaultHighlightIcon = $q.dark.isActive
		? securityHighlightIcon
		: securityHighlightDarkIcon;
	return [
		{
			key: 'account',
			label: t('account'),
			img: accountDefaultIcon,
			activeImg: accountDefaultHighlightIcon,
			link: '/options/account'
		},
		{
			key: 'appearance',
			label: t('Appearance'),
			img: appearanceDefaultIcon,
			activeImg: appearanceDefaultHighlightIcon,
			link: '/options/appearance'
		},
		// {
		// 	key: 'collect',
		// 	label: t('Collect'),
		// 	img: collectDefaultIcon,
		// 	activeImg: collectDefaultHighlightIcon,
		// 	link: '/options/collect'
		// },
		// {
		// 	key: 'translate',
		// 	label: t('Translate'),
		// 	img: translateDefaultIcon,
		// 	activeImg: translateDefaultHighlightIcon,
		// 	link: '/options/translate'
		// },
		{
			key: 'security',
			label: t('security'),
			img: securityDefaultIcon,
			activeImg: securityDefaultHighlightIcon,
			link: '/options/security'
		}
	];
});
const initOption = computed(() => [
	{
		key: 'sub1',
		children: options.value
	}
]);

const router = useRouter();
const route = useRoute();
const active = ref(options.value[0].key);
const userData = ref();
const scrollRef = ref();

const LarePassTextIcon = computed(() => {
	return $q.dark.isActive ? LarePassTextDarkIcon : LarePassTextIconLight;
});
const currentItem = computed(
	() => options.value.find((item) => item.key === active.value) || { label: '' }
);

const selectHandler = (data: any) => {
	router.push({
		path: data.item.link
	});
};

const menuActive = () => {
	const link = route.path;
	const target = options.value.find((item) => item.link === link);
	if (target) {
		active.value = target.key;
		scrollRef.value &&
			scrollRef.value?.$refs?.scrollRef?.setScrollPosition('vertical', 0, 10);
	}
};

watch(
	() => route.path,
	() => {
		menuActive();
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.plugin-options-layout {
}
.bex-options-scroll-wrapper {
	height: calc(100vh);
}
.title {
	color: #1f1814;
	font-size: 24px;
	font-weight: 700;
	line-height: 32px;
	padding: 20px 32px 0 32px;
}
::v-deep(.my-active-link) {
	color: $light-blue-default;
	background: $light-blue-soft;
}
::v-deep(.q-drawer.q-drawer--bordered) {
	border-color: $separator-2;
}
</style>
