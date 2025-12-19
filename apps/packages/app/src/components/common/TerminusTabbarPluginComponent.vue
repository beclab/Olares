<template>
	<div
		class="tab-container bg-background-6 column justify-between no-wrap items-center flex-gap-md q-py-md"
		:style="{
			width: show ? '64px' : '44px'
		}"
	>
		<div class="column items-center no-wrap flex-gap-md" style="flex: 1">
			<div
				class="full-width text-center cursor-pointer"
				:class="{
					'action-rotate': show
				}"
				@click="toggleMenu"
			>
				<q-img :src="actionIcon" :ratio="1" spinner-size="0px" width="20px" />
			</div>
			<div
				class="column items-center justify-between no-wrap flex-gap-xl menu-list-top"
				style="flex: 1"
			>
				<div class="column flex-gap-md">
					<div
						v-for="(item, index) in itemList"
						:key="item.identify"
						@click="updateCurrent(index)"
						class="column items-center relative-position tab-item"
					>
						<div
							class="tab-icon-wrapper"
							:class="{
								'bg-background-hover': current === index
							}"
						>
							<div
								v-if="current !== index"
								class="tab-icon-size tab-icon-size-hover"
							>
								<img
									class="img-normal"
									:src="
										getRequireImage(
											`tabs/${item.normalImage}${
												$q.dark.isActive ? '-dark' : ''
											}.svg`
										)
									"
								/>
								<img
									class="img-hover"
									:src="
										getRequireImage(
											`tabs/${item.hoverImage}${
												$q.dark.isActive ? '-dark' : ''
											}.svg`
										)
									"
								/>
							</div>
							<div v-if="current == index" class="tab-icon-size">
								<img
									:src="
										getRequireImage(
											$q.dark.isActive
												? `tabs/${item.activeImage}-dark.svg`
												: `tabs/${item.activeImage}.svg`
										)
									"
								/>
							</div>

							<div
								class="trans-num text-caption text-white row items-center"
								:class="{
									'trans-num-active': current == index
								}"
								v-if="item.badge && item.badge.length > 0"
							>
								{{ item.badge }}
							</div>
						</div>

						<div
							class="text-overline tab-label"
							:class="current !== index ? 'text-ink-3' : 'text-ink-1'"
							:style="{
								height: show ? '12px' : '0px',
								marginTop: show ? '4px' : '0px',
								overflow: 'hidden'
							}"
						>
							{{ t(`bex.${item.name}`) }}
						</div>

						<bt-tooltip
							v-if="!show"
							:label="t(`bex.${item.name}`)"
							anchor="center left"
							self="center end"
							:offset="[4, 0]"
						/>
					</div>
				</div>
			</div>
		</div>
		<div class="column items-center no-wrap">
			<slot name="footer"></slot>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../utils/imageUtils';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useFilesStore } from '../../stores/files';
import { useTermipassStore } from '../../stores/termipass';
import menuActionIcon from 'src/assets/plugin/menu-action.svg';
import menuActionIconDark from 'src/assets/plugin/menu-action-dark.svg';
import BtTooltip from 'src/components/base/BtTooltip.vue';
import { useUserStore } from 'src/stores/user';

import { computed, ref } from 'vue';
import { useQuasar } from 'quasar';
import { tabsIgnore } from 'src/platform/interface/bex/front/bexTabOptions';

const props = defineProps({
	current: {
		type: Number,
		default: 0,
		required: true
	}
});
const userStore = useUserStore();
const show = ref(false);
const emit = defineEmits(['updateCurrent']);

const $router = useRouter();
const $q = useQuasar();

const itemList = computed(() => {
	if (!userStore.current_user?.isLargeVersion12 && process.env.IS_BEX) {
		return termipassStore.tabItems.filter(
			(item) => !tabsIgnore.includes(item.identify)
		);
	} else {
		return termipassStore.tabItems;
	}
});

const actionIcon = computed(() => {
	return $q.dark.isActive ? menuActionIconDark : menuActionIcon;
});

const termipassStore = useTermipassStore();

const tabItemOption = computed(() => {
	const top = termipassStore.tabItems.slice(0, -1);
	const bottom = termipassStore.tabItems.slice(-1);
	return {
		top,
		bottom
	};
});

const { t } = useI18n();

const filesStore = useFilesStore();

const updateCurrent = (index: number) => {
	filesStore.previousStack = {};
	if (index === props.current) {
		return;
	}
	if (termipassStore.tabItems[index].tabChanged) {
		const result = termipassStore.tabItems[index].tabChanged();
		if (result) {
			return;
		}
	}
	if (termipassStore.tabItems[index].to) {
		$router.replace(termipassStore.tabItems[index].to);
		return;
	}
	emit('updateCurrent', index);
};

const toggleMenu = () => {
	show.value = !show.value;
};
</script>

<style scoped lang="scss">
.tab-container {
	min-width: 44px;
	height: 100vh;
	border-left: 1px solid $separator-2;
	transition: width 0.15s ease-in-out;
	.action-rotate {
		transform: rotate(180deg);
		transform-origin: center;
		transition: all 0.15s ease-in-out;
	}
	.menu-list-top {
		margin-top: 6px;
	}
	.tab-item {
		font-size: 0px;
	}
	.tab-icon-wrapper {
		border-radius: 8px;
		&:hover {
			background: $background-hover;
		}
		.tab-icon-size {
			width: 32px;
			height: 32px;
			display: flex;
			justify-content: center;
			align-items: center;
			.img-hover {
				display: none;
			}
			&.tab-icon-size-hover:hover {
				.img-normal {
					display: none;
				}
				.img-hover {
					display: inline-block;
				}
			}
		}
	}
	.tab-label {
		transition: all 0.15s ease-in-out;
	}
}
</style>
