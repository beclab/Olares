<template>
	<div class="relative-position">
		<MenuHeader size="xxl">
			<template #title-after>
				<q-tabs
					class="q-pl-lg q-ml-md text-ink-2"
					content-class="tabs-content-wrapper"
					style="font-size: 16px"
					v-model="tab"
					align="left"
					active-color="blue-6"
					:breakpoint="0"
					no-caps
					narrow-indicator
				>
					<q-tab :ripple="false" content-class="tabs-content-wrapper" :name="1">
						{{ $t('POD_LIST') }}
					</q-tab>
					<q-tab :ripple="false" :name="2">{{ $t('RESOURCE') }}</q-tab>
				</q-tabs>
			</template>
		</MenuHeader>
		<bt-scroll-area
			style="height: calc(100vh - 56px); width: 100%"
			:content-style="contentStyle"
			:content-active-style="contentStyle"
		>
			<div class="relative-position q-mx-xxl q-mb-lg">
				<PodList v-if="tab === 1"></PodList>
				<ClusterResource v-if="tab === 2"></ClusterResource>
			</div>
		</bt-scroll-area>
		<RouterViewTransition></RouterViewTransition>
	</div>
</template>

<script setup lang="ts">
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import { ref } from 'vue';
import PodList from './PodList.vue';
import ClusterResource from '@apps/control-hub/src/pages/NodeResource/IndexPage.vue';
import RouterViewTransition from '@apps/control-panel-common/src/components/RouterViewTransition.vue';

const tab = ref(1);

const contentStyle = {
	minWidth: 'auto',
	width: '100%'
};
</script>
