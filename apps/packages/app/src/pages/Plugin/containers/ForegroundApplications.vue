<template>
	<div class="front-apps-container">
		<VueDraggable
			v-model="list"
			colWidth="55px"
			gapRow="md"
			gapCol="none"
			class="row front-apps flex-gap-y-md"
			ghostClass="ghost"
			group="app-drag-group"
		>
			<div
				v-for="item in list"
				:key="item.id"
				class="app-item column items-center justify-center"
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
		</VueDraggable>
	</div>
</template>

<script setup lang="ts">
import appDeleteIcon from 'src/assets/plugin/app-delete.svg';
import { useAppsStore } from 'src/stores/bex/apps';
import ShakeDom from 'src/pages/Plugin/components/ShakeDom.vue';
import { VueDraggable } from 'vue-draggable-plus';
import { computed, onMounted, ref } from 'vue';

const appsStore = useAppsStore();
const edited = ref(false);

interface Props {
	showAction?: boolean;
}

const props = defineProps<Props>();
const emits = defineEmits(['updateOne']);

const openUrl = (url: string) => {
	if (!props.showAction) {
		appsStore.openUrl(url);
	}
};

const list = computed({
	get: () => appsStore.foregroundApps,
	set: (value) => {
		const ids = value.map((item) => item.id);
		appsStore.sortApp(ids);
		if (!edited.value) {
			edited.value = true;
			emits('updateOne');
		}
	}
});

const reset = () => {
	edited.value = false;
};
onMounted(() => {
	edited.value = false;
});

defineExpose({ reset });
</script>

<style lang="scss" scoped>
.front-apps-container {
	.app-item {
		height: 64px;
		padding: 4px 0;
		width: 55px;
	}
	.front-apps {
		.app-icon {
			overflow: hidden;
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
	::v-deep(.ghost .app-icon) {
		border-radius: 9px;
	}
}
</style>
