<template>
	<div class="vc-manager-root">
		<terminus-title-bar title="History" />
		<terminus-scroll-area class="vc-scroll">
			<template v-slot:content>
				<div
					class="q-pa-md example-row-equal-width"
					v-for="(history, index) in historys"
					:key="index"
				>
					<div class="row">
						<div class="col">
							{{
								history.type == 'check' ? history.checkItem : history.reasonDesc
							}}
						</div>
						<div class="col">
							{{ date.formatDate(history.date, 'HH:mm:ss') }}
						</div>
						<div
							class="col"
							v-if="
								history.type == 'check' &&
								history.result &&
								history.result.before
							"
						>
							{{ history.result.before }}
						</div>
						<div
							class="col"
							v-if="
								history.type == 'check' &&
								history.result &&
								history.result.after
							"
						>
							{{ history.result.after }}
						</div>
						<div class="col" v-if="history.type == 'check' && history.result">
							{{ history.result.description }}
						</div>
					</div>
				</div>
			</template>
		</terminus-scroll-area>
	</div>
</template>
<script setup lang="ts">
import { ref } from 'vue';
import { useTermipassStore } from '../../../stores/termipass';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import TerminusScrollArea from '../../../components/common/TerminusScrollArea.vue';
import { date } from 'quasar';

const termipassStore = useTermipassStore();
const historys = ref(termipassStore.state.publicActions.getCheckHistory());
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
