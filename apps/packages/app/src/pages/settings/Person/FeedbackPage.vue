<template>
	<page-title-component
		:show-back="true"
		:title="t('Communication and feedback')"
	/>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<bt-list
			:first="index === 0"
			v-for="(item, index) in linkArray"
			:key="item.link"
		>
			<bt-form-item
				:margin-top="false"
				:chevron-right="false"
				:width-separator="false"
			>
				<template v-slot:all>
					<div
						class="row justify-between items-center full-width q-px-lg cursor-pointer"
						@click="openNewWindow(item.link)"
					>
						<div class="row justify-start items-center">
							<q-img style="height: 40px; width: 40px" :src="item.icon" />
							<div class="column q-px-lg">
								<div class="text-subtitle1 text-ink-1">{{ item.title }}</div>
								<div class="text-body3 text-ink-3">
									{{ item.link }}
								</div>
							</div>
						</div>
						<q-icon class="text-ink-1" name="sym_r_open_in_new" size="24px" />
					</div>
				</template>
			</bt-form-item>
		</bt-list>
	</bt-scroll-area>
</template>
<script setup lang="ts">
import PageTitleComponent from '../../../components/settings/PageTitleComponent.vue';
import BtFormItem from '../../../components/settings/base/BtFormItem.vue';
import BtList from 'src/components/settings/base/BtList.vue';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';

const { t } = useI18n();

interface LinkProps {
	title: string;
	icon: string;
	link: string;
}

const linkArray = ref<LinkProps[]>([
	{
		title: 'Github',
		icon: 'settings/plant/Github.svg',
		link: 'https://github.com/beclab/Olares'
	},
	{
		title: 'Youtube',
		icon: 'settings/plant/Youtube.svg',
		link: 'https://www.youtube.com/@OlaresOS/featured'
	},
	{
		title: 'Discord',
		icon: 'settings/plant/Discord.svg',
		link: 'https://discord.com/invite/BzfqrgQPDK'
	},
	{
		title: 'Twitter',
		icon: 'settings/plant/Twitter.svg',
		link: 'https://x.com/Olares_OS'
	},
	{
		title: 'Facebook',
		icon: 'settings/plant/Facebook.svg',
		link: 'https://www.facebook.com/people/Olares/61579156063357/'
	}
]);

const openNewWindow = (url: string) => {
	window.open(url, '_blank');
};
</script>

<style scoped lang="scss"></style>
