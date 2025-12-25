<template>
	<div class="tabs-root">
		<div class="tabs-content">
			<div class="tab-separator" style="width: 100%; height: 1px"></div>
			<div class="items-bg" />
		</div>

		<div class="tabs-items row items-center justify-evenly">
			<div
				v-for="(item, index) in termipassStore.tabItems"
				:key="item.identify"
				:style="`width:calc((100% - 20px)/${termipassStore.tabItems.length})`"
				@click="updateCurrent(index)"
				class="tab-item column justify-end items-center"
			>
				<div v-if="current !== index">
					<img
						:src="getRequireImage(`tabs/${item.normalImage}.svg`)"
						class="tab-icon-size"
					/>
				</div>
				<div v-else class="tab-icon-active-bg"></div>

				<div
					class="tab-title-base text-body3"
					:class="current !== index ? 'text-grey-6' : 'text-ink-1'"
				>
					{{ t(item.name) }}
				</div>

				<div class="active-circle-bg" v-if="current == index"></div>
				<div
					class="row items-center justify-center active-icon"
					v-if="current == index"
				>
					<img
						:src="
							getRequireImage(
								$q.dark.isActive && item.darkActiveImage
									? `tabs/${item.darkActiveImage}.svg`
									: `tabs/${item.activeImage}.svg`
							)
						"
						class="tab-icon-size"
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
		</div>

		<div class="safe-area-bottom"></div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../utils/imageUtils';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { useFilesStore } from './../../stores/files';
import { useTermipassStore } from '../../stores/termipass';

const props = defineProps({
	current: {
		type: Number,
		default: 0,
		required: true
	}
});

const emit = defineEmits(['updateCurrent']);

const $router = useRouter();

const termipassStore = useTermipassStore();

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
</script>

<style scoped lang="scss">
.tabs-root {
	width: 100%;

	.tabs-content {
		height: 80px;
		width: 100%;
		padding-top: 16px;
		.tab-separator {
			background-color: $separator;
		}

		.items-bg {
			width: 100%;
			height: 64px;
			background-color: $background-1;
		}
	}

	.tabs-items {
		height: 85px;
		width: 100%;
		margin-top: -85px;

		.tab-item {
			height: 100%;
			position: relative;

			.tab-title-base {
				line-height: 14px;
				text-align: center;
				margin-bottom: 5px;
			}

			.tab-icon-size {
				height: 24px;
				width: 24px;
			}

			.trans-num {
				position: absolute;
				display: inline-block;
				background-color: $negative;
				height: 16px;
				border-radius: 8px;
				padding: 1px 4px;
				top: 30px;
				left: 50%;
				border: 1px solid $background-1;
			}

			.trans-num-active {
				top: 20px;
			}

			.tab-icon-active-bg {
				margin-top: 5px;
				border-radius: 30px;
				background-color: $background-1;
				width: 60px;
				height: 60px;
				border-color: $separator;
				border-width: 1px;
				border-style: solid;
				border-top: 50%;
			}

			.active-circle-bg {
				position: absolute;
				top: 22px;
				width: 65px;
				height: 44px;
				background-color: $background-1;
			}

			.active-icon {
				width: 46px;
				height: 46px;
				position: absolute;
				top: 17px;
				border-radius: 23px;
				border-width: 1px;
				border-color: $separator;
				border-style: solid;
			}
		}
	}

	.safe-area-bottom {
		width: 100%;
		height: calc(env(safe-area-inset-bottom));
	}
}
</style>
