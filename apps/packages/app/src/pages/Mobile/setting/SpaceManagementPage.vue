<template>
	<div class="vc-manager-root">
		<terminus-title-bar :title="t('olares_space_management')" />
		<terminus-scroll-area class="vc-scroll">
			<template v-slot:content>
				<!-- <q-list></q-list> -->
				<div class="column text-ink-2 items-center" style="margin-top: 180px">
					<img src="../../../assets/layout/nodata.svg" />
					<span class="q-mb-md text-body2" style="margin-top: 32px">
						{{ t('there_is_nothing_for_now') }}
					</span>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>

<script setup lang="ts">
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<style scoped lang="scss">
.vc-manager-root {
	width: 100%;
	height: 100%;
}

.vc-scroll {
	width: 100%;
	height: calc(100% - 56px);
}
</style>
