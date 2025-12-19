<template>
	<PageCard :title="$t('bex.translate')">
		<div class="column no-wrap flex-gap-y-md">
			<div>
				<div class="text-body3 text-ink-3">
					{{ $t('bex.translate_service') }}
				</div>
				<div class="q-mt-xs">
					<BtSelect
						v-model="rule.translator"
						:options="options"
						:border="true"
						:height="40"
						:iconSize="24"
						classes="q-px-md"
						menuClasses="q-pa-xs"
						:menuItemHeight="40"
						@update:model-value="translateHandler2"
					></BtSelect>
				</div>
			</div>
			<div>
				<div class="row items-center justify-between">
					<span class="text-Body3 text-ink-1">
						<span>{{ $t('bex.translate_page_always') }}</span>
						<q-icon
							name="sym_r_keyboard_arrow_down"
							size="16px"
							color="ink-1"
							class="q-ml-xs"
						/>
					</span>
					<bt-switch
						v-model="rule.transOpen"
						truthyTrackColor="light-blue-default"
						size="xs"
						class="custom-toggle-wrapper"
						@update:model-value="translateHandler"
					/>
				</div>
			</div>
			<CustomButton
				color="yellow-default"
				class="full-width"
				@click="handleTransToggle"
			>
				<template #label>
					<div class="row items-center no-wrap">
						<span class="relative-position">
							<q-icon name="translate" size="20px" />
							<img
								:src="checkedIcon"
								alt="checked"
								style="width: 8px; height: 8px"
								class="absolute-bottom-right z-top"
								v-show="transOpen"
							/>
						</span>
						<span class="q-ml-sm">{{
							transOpen ? $t('bex.show_original') : $t('bex.translate')
						}}</span>
					</div>
				</template>
			</CustomButton>
		</div>

		<!-- <q-btn
			color="primary"
			icon="check"
			label="save Rule"
			@click="handleSaveRule"
		/> -->
		<q-inner-loading :showing="loading"> </q-inner-loading>
	</PageCard>
</template>

<script setup lang="ts">
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import BtSelect from 'src/components/base/BtSelect.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useTranslate } from 'src/composables/mobile/useTranslate';
import { onMounted, onUnmounted } from 'vue';
import {
	bexFrontBusOff,
	bexFrontBusOn
} from 'src/platform/interface/bex/utils';
import checkedIcon from 'src/assets/plugin/checked.svg';

const {
	handleTransToggle,
	translator,
	options,
	translateHandler2,
	transOpen,
	translateHandler,
	loading,
	getTransRule,
	rule
} = useTranslate();

onMounted(() => {
	getTransRule();
	bexFrontBusOn('COLLECTION_TAB_UPDATE', getTransRule);
});

onUnmounted(() => {
	bexFrontBusOff('COLLECTION_TAB_UPDATE', getTransRule);
});
</script>

<style lang="scss" scoped>
.custom-toggle-wrapper {
	::v-deep(.q-toggle__inner--truthy .q-toggle__thumb:after) {
		background-color: $ink-on-brand !important;
	}
}
</style>
