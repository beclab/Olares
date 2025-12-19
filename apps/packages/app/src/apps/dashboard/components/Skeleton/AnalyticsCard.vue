<template>
	<template v-if="showing">
		<MyCard v-for="item in 2" :key="item">
			<div class="q-gutter-y-xl">
				<div class="row justify-between">
					<div class="row items-center">
						<div class="q-mr-md">
							<MyImg src="" :loading="loading"></MyImg>
						</div>
						<div class="text-h4 text-grey-10">
							<q-skeleton width="80px" />
						</div>
					</div>
					<QButtonStyle>
						<q-skeleton width="32px" height="32px" />
					</QButtonStyle>
				</div>
				<Chart />
				<MyGridLayout col-width="135px">
					<CardInfo v-for="item in 3" :key="item" />
				</MyGridLayout>
			</div>
		</MyCard>
	</template>
</template>
<script setup lang="ts">
import MyCard from '@apps/dashboard/src/components/MyCard.vue';
import Chart from './ChartBase.vue';
import CardInfo from './CardInfo2.vue';
import MyGridLayout from '@apps/control-panel-common/src/components/MyGridLayout.vue';
import MyImg from '@apps/control-panel-common/src/components/MyImg.vue';
import { computed } from 'vue';

interface Props {
	showing?: boolean;
}

const props = withDefaults(defineProps<Props>(), {});

const loading = computed(() => props.showing);
</script>
