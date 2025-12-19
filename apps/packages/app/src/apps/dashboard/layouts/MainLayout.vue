<template>
	<q-layout view="hHh lpR fFf">
		<q-drawer show-if-above :width="240" side="left" bordered :breakpoint="0">
			<bt-menu
				active-class="my-active-link"
				:items="items"
				v-model="active"
				style="width: 239px"
				@select="selectHandler"
			/>
		</q-drawer>
		<q-page-container class="relative-position dashboard-container">
			<bt-scroll-area class="dashboard-scroll-wrapper" ref="scrollRef">
				<div
					class="text-h3 q-px-xxl q-mb-md text-ink-1"
					style="margin-top: 56px"
				>
					{{ currentItem.label }}
				</div>
				<div class="q-px-xxl q-pb-xl q-pt-md" style="min-width: 800px">
					<router-view />
				</div>
			</bt-scroll-area>
		</q-page-container>
	</q-layout>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useAppDetailStore } from '@apps/dashboard/src/stores/AppDetail';
import { useI18n } from 'vue-i18n';

const AppDetailStore = useAppDetailStore();
const { t } = useI18n();
const options = [
	{
		key: 'overview',
		label: t('OVERVIEW'),
		icon: 'sym_r_dashboard',
		link: '/overview'
	},
	{
		key: 'applications',
		label: t('APPLICATIONS'),
		icon: 'sym_r_wysiwyg',
		link: '/applications'
	}
	// {
	// 	key: 'analytics',
	// 	label: t('ANALYTICS'),
	// 	icon: 'sym_r_leaderboard',
	// 	link: '/analytics'
	// }
];
const initOption = {
	key: 'sub1',
	label: `${AppDetailStore.user.username}’s Olares`,
	children: options
};
const items = ref([initOption]);

const router = useRouter();
const route = useRoute();
const active = ref(options[0].key);
const userData = ref();
const scrollRef = ref();

const currentItem = computed(
	() => options.find((item) => item.key === active.value) || { label: '' }
);

const selectHandler = (data: any) => {
	router.push({
		path: data.item.link
	});
};

const menuActive = () => {
	const link = route.path.split('/')[1];
	const target = options.find((item) => item.link === `/${link}`);
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
.dashboard-scroll-wrapper {
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
	border-color: $separator;
}
</style>
