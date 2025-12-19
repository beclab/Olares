<template>
	<q-layout class="control-hub-layout" view="hHh lpR fFf">
		<q-drawer show-if-above :width="240" side="left" bordered :breakpoint="0">
			<bt-scroll-area class="full-height" style="padding-top: 6px">
				<bt-menu
					active-class="my-active-link"
					:items="items"
					v-model="active"
					@select="selectHandler"
				/>
			</bt-scroll-area>
		</q-drawer>
		<q-page-container class="settings_content_view">
			<div class="layout-content-wrapper">
				<div>
					<router-view />
					<TerminalBox></TerminalBox>
				</div>
			</div>
		</q-page-container>
	</q-layout>
</template>

<script lang="ts" setup>
import { onMounted, computed, ref } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useSplitMenu } from '@apps/control-panel-common/src/stores/menu';
import { useI18n } from 'vue-i18n';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
import {
	updateBreadcrumbs,
	options,
	currentItem,
	active,
	options2,
	options3,
	options4,
	updateKey
} from './breadcrumbs';
import TerminalBox from '@apps/control-hub/src/pages/Terminal/TerminalBox.vue';
import { useTerminalStore } from '../stores/TerminalStore';
import { watch } from 'vue';
import { onBeforeMount } from 'vue';
import { componentName } from '@apps/control-hub/src/router/const';
const terminalStore = useTerminalStore();
const { t } = useI18n();
const appDetailStore = useAppDetailStore();
const splitMenu = useSplitMenu();

const items = computed(() => {
	const item1 = [
		{
			key: 'terminus',
			label: t('TERMINUS'),
			children: options.value
		},
		{
			key: 'resource',
			label: t('RESOURCE'),
			children: options2.value
		}
	];
	const item2 = computed(() => {
		return options3.value.length > 0
			? [
					{
						key: 'middleware',
						label: t('MIDDLEWARE'),
						children: options3.value
					}
			  ]
			: [];
	});

	const item3 = [
		{
			key: 'terminal',
			label: t('TERMINAL'),
			children: options4.value
		}
	];

	return appDetailStore.isDemo
		? item1
		: appDetailStore.isAdmin
		? item1.concat(item2.value).concat(item3)
		: item1.concat(item2.value);
});

const optionsAll = [
	...options.value,
	...options2.value,
	...options3.value,
	...options4.value
];

const router = useRouter();
const route = useRoute();

const selectHandler = (data: any) => {
	splitMenu.changeStatus(data.item.key);
	const breadcrumbsData = {
		title: data.item.label,
		icon: data.item.icon
	};
	updateBreadcrumbs(breadcrumbsData, true);
	route.meta.item = data.item.label;
	router.push({
		path: data.item.link
	});
};

const init = () => {
	const link = route.path.split('/')[1];
	let target: any = optionsAll[0];
	if (link === 'site-middleware') {
		target = optionsAll.find((item) => item.link === route.path);
	} else if (link === 'terminal') {
		target = optionsAll.find((item) => item.key === link);
		if (route.params.node) {
			terminalStore.setCurrentNode(route.params.node);
		}
	} else {
		target = optionsAll.find((item) => item.link === `/${link}`);
	}
	if (target) {
		updateKey(target.key);
		const breadcrumbsData = {
			title: target.label,
			icon: target.icon
		};
		updateBreadcrumbs(breadcrumbsData, true);
	}
};

watch(
	() => route.fullPath,
	() => {
		if (route?.name === componentName.TERMINAL_APP) {
			terminalStore.terminalPanelInit();
			terminalStore.terminalPanelStatus(true);
		} else {
			terminalStore.terminalPanelStatus(false);
		}
	},
	{
		immediate: true
	}
);

onBeforeMount(() => {
	init();
});
</script>

<style lang="scss" scoped>
.layout-content-wrapper {
	// background: $bg-content;
	transform: translate(0);
}
.control-hub-sidebar {
	border: 1px solid red;
}
.title {
	color: #1f1814;
	font-size: 24px;
	font-weight: 700;
	line-height: 32px;
	padding: 20px 32px 0 32px;
}

::v-deep(.my-active-link) {
	color: $blue-6;
	background: $blue-soft;
}
::v-deep(.q-drawer.q-drawer--bordered) {
	border-color: $separator;
}
</style>
<style></style>
