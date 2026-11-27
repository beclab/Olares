<template>
	<div class="feedback-page fit column">
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
							class="feedback-row full-width q-px-lg cursor-pointer"
							@click="openNewWindow(item.link)"
						>
							<q-img
								class="feedback-row__icon"
								style="height: 40px; width: 40px"
								:src="item.icon"
							/>
							<div class="feedback-row__text">
								<div class="text-subtitle1 text-ink-1 feedback-ellipsis">
									{{ item.title }}
								</div>
								<div class="text-body3 text-ink-3 feedback-ellipsis">
									{{ item.link }}
								</div>
							</div>
							<q-icon
								class="feedback-row__open text-ink-1"
								name="sym_r_open_in_new"
								size="24px"
							/>
						</div>
					</template>
				</bt-form-item>
			</bt-list>
		</bt-scroll-area>
	</div>
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
		title: 'GitHub',
		icon: 'settings/plant/Github.svg',
		link: 'https://github.com/beclab/Olares'
	},
	{
		title: 'YouTube',
		icon: 'settings/plant/Youtube.svg',
		link: 'https://www.youtube.com/@OlaresOS/featured'
	},
	{
		title: 'Discord',
		icon: 'settings/plant/Discord.svg',
		link: 'https://discord.com/invite/BzfqrgQPDK'
	},
	{
		title: 'X',
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

<!--
	Use a non-scoped (global) style scoped under `.feedback-page` so it can
	reach into the BtList / BtFormItem child components. Every ancestor up the
	chain must be allowed to shrink below its content width (min-width: 0), and
	the row itself uses a CSS grid with a `minmax(0, 1fr)` middle track so the
	link column is bounded by the available width and truncates with an
	ellipsis on any screen size instead of pushing the page wider.
-->
<style lang="scss">
.feedback-page {
	.q-scrollarea__content {
		max-width: 100%;
	}

	.form-item-root {
		min-width: 0;
		max-width: 100%;
		overflow: hidden;
	}

	.form-item {
		min-width: 0;
		width: 100%;
	}

	.feedback-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		column-gap: 16px;
	}

	.feedback-row__text {
		min-width: 0;
	}

	.feedback-ellipsis {
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}
</style>
