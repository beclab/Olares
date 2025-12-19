<template>
	<div class="recommend-page column">
		<title-bar>
			<template v-slot:before>
				<bt-breadcrumbs
					:title="t('main.recommendations')"
					icon="sym_r_featured_play_list"
					:breadcrumb="tab !== 'group'"
				>
					<template v-slot:breadcrumb>
						<div @click="onBackClick" class="row items-center cursor-pointer">
							<q-icon size="20px" name="sym_r_arrow_back_ios_new" />
						</div>
						<q-breadcrumbs-el
							class="left-align text-h6 text-ink-1 cursor-pointer"
							:label="t('main.recommendations')"
							@click="
								argoStore.cronLabel = '';
								argoStore.namespace = '';
								argoStore.workflow_id = '';
							"
						/>
						<q-breadcrumbs-el
							class="left-align text-h6 text-ink-1 cursor-pointer"
							v-if="argoStore.cronLabel"
							:label="argoStore.cronLabel"
							@click="argoStore.workflow_id = ''"
						/>
						<q-breadcrumbs-el
							class="left-align text-h6 text-ink-1 cursor-pointer"
							v-if="argoStore.workflow_id"
							:label="argoStore.workflow_id"
						/>
					</template>
				</bt-breadcrumbs>
			</template>
		</title-bar>

		<q-tab-panels
			v-model="tab"
			animated
			keep-alive
			class="recommend-tab-panels"
		>
			<q-tab-panel name="group" class="recommend-tab-panel">
				<workflow-group />
			</q-tab-panel>

			<q-tab-panel name="list" class="recommend-tab-panel">
				<workflow-list />
			</q-tab-panel>

			<q-tab-panel name="detail" class="recommend-tab-panel">
				<workflow-detail />
			</q-tab-panel>
		</q-tab-panels>
	</div>
</template>

<script setup lang="ts">
import { useArgoStore } from '../../../stores/argo';
import { computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import BtBreadcrumbs from '../../../components/base/BtBreadcrumbs.vue';
import TitleBar from '../../../components/rss/TitleBar.vue';
import WorkflowGroup from './recommend/WorkflowGroup.vue';
import WorkflowList from './recommend/WorkflowList.vue';
import WorkflowDetail from './recommend/WorkflowDetail.vue';

const { t } = useI18n();
const argoStore = useArgoStore();

const tab = computed(() => {
	if (argoStore.workflow_id) {
		return 'detail';
	}
	if (argoStore.cronLabel && argoStore.namespace) {
		return 'list';
	}
	return 'group';
});

const onBackClick = () => {
	if (argoStore.workflow_id) {
		argoStore.workflow_id = '';
		return;
	}
	argoStore.namespace = '';
	argoStore.cronLabel = '';
};

onMounted(() => {
	argoStore.get_cron_workflows();
});
</script>

<style scoped lang="scss">
.recommend-page {
	height: 100%;
	width: 100%;

	.recommend-tab-panels {
		height: calc(100% - 56px);

		.recommend-tab-panel {
			padding: 0;
		}
	}
}
</style>
