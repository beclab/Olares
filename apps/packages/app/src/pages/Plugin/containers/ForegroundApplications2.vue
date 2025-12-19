<template>
	<div class="front-apps-container relative-position">
		<div class="apps-scroll-wrapper">
			<StepScroll>
				<div class="row no-wrap front-apps q-pa-sm">
					<div
						v-for="item in appsStore.foregroundApps"
						:key="item.id"
						class="app-item column items-center"
					>
						<div
							class="relative-position"
							:class="{
								'cursor-pointer': !showAction
							}"
						>
							<ShakeDom :shake="showAction">
								<q-img
									class="app-icon"
									:src="item.icon"
									:ratio="1"
									width="32px"
									spinner-size="0"
									@click="openUrl(item.url)"
								>
									<template #loading>
										<q-skeleton
											type="rect"
											square
											width="32px"
											height="32px"
											animation="fade"
										/>
									</template>
								</q-img>
								<q-img
									v-show="showAction"
									class="icon"
									:src="appDeleteIcon"
									:ratio="1"
									width="16px"
									spinner-size="0px"
									@click="appsStore.deleteApp(item.id)"
								/>
							</ShakeDom>
						</div>
						<div
							class="text-center text-caption text-ink-2 full-width ellipsis q-mt-sm"
						>
							{{
								$te(`system_app.${item.title}`)
									? $t(`system_app.${item.title}`)
									: item.title
							}}
						</div>
					</div>
				</div>
			</StepScroll>
		</div>
	</div>
</template>

<script setup lang="ts">
import appDeleteIcon from 'src/assets/plugin/app-delete.svg';
import { useAppsStore } from 'src/stores/bex/apps';
import ShakeDom from 'src/pages/Plugin/components/ShakeDom.vue';
import StepScroll from 'src/pages/Plugin/components/StepScroll.vue';

const appsStore = useAppsStore();

interface Props {
	showAction?: boolean;
}

const props = defineProps<Props>();

const openUrl = (url: string) => {
	if (!props.showAction) {
		appsStore.openUrl(url);
	}
};
</script>

<style lang="scss" scoped>
.front-apps-container {
	position: relative;
	width: 100%;
	height: 82px;
	.apps-scroll-wrapper {
		position: absolute;
		left: 0;
		right: 0;
		top: 50%;
		transform: translateY(-50%);
	}
	.front-apps {
		.app-icon {
			border-radius: 9px;
		}
		.icon {
			position: absolute;
			top: 0px;
			right: 0px;
			transform: translate(8px, -6px);
			cursor: pointer;
		}
	}
	.app-item {
		width: 55px;
	}
}
</style>
