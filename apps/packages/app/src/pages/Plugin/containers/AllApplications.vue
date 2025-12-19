<template>
	<div class="all-apps-container">
		<VueDraggable
			v-model="list"
			colWidth="55px"
			gapRow="md"
			gapCol="none"
			class="row front-apps flex-gap-y-md"
			:contentLength="list.length"
			ghostClass="ghost"
			group="app-drag-group"
		>
			<div
				v-for="item in list"
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
							:src="actionIcon(item.id)"
							:ratio="1"
							width="16px"
							spinner-size="0px"
							@click="actionHandler(item.id)"
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
		</VueDraggable>
	</div>
</template>

<script setup lang="ts">
import MyGridLayout from '@apps/control-panel-common/components/MyGridLayout.vue';
import appDeleteIcon from 'src/assets/plugin/app-delete.svg';
import appAddIcon from 'src/assets/plugin/app-add.svg';
import { useAppsStore } from 'src/stores/bex/apps';
import { computed, onMounted } from 'vue';
import ShakeDom from 'src/pages/Plugin/components/ShakeDom.vue';
import { vDraggable, VueDraggable } from 'vue-draggable-plus';
import { difference } from 'lodash';

interface Props {
	showAction?: boolean;
}

const props = defineProps<Props>();

const appsStore = useAppsStore();

const allApps = computed(() =>
	appsStore.foregroundApps.concat(appsStore.backgroundApps)
);

const actionIcon = (id: string) => {
	const target = appsStore.foregroundApps.find((item) => item.id === id);
	return target ? appDeleteIcon : appAddIcon;
};

const actionHandler = (id: string) => {
	const target = appsStore.foregroundApps.find((item) => item.id === id);
	if (target) {
		appsStore.deleteApp(id);
	} else {
		appsStore.addApp(id);
	}
};

const openUrl = (url: string) => {
	if (!props.showAction) {
		appsStore.openUrl(url);
	}
};

const list = computed({
	get: () => allApps.value,
	set: (value) => {
		const ids = value.map((item) => item.id);
		console.log(ids);
	}
});

onMounted(() => {
	appsStore.init();
});
</script>

<style lang="scss" scoped>
.all-apps-container {
	.app-item {
		height: 64px;
		padding: 4px 0;
		width: 55px;
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
}
</style>
