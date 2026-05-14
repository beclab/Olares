<template>
	<PageCard :title="$t('bex.translate')">
		<div class="column no-wrap flex-gap-y-md">
			<!-- Language Selectors Row -->
			<div class="row items-center" style="gap: 8px">
				<!-- Source Language Selector -->
				<div class="language-selector-wrapper">
					<div
						class="language-selector"
						:class="{ 'selector-disabled': !contentScriptReady }"
						@click.stop="
							contentScriptReady && (showFromLangMenu = !showFromLangMenu)
						"
					>
						<div class="language-selector-content">
							<div class="language-selector-text">
								<div class="language-name text-subtitle3 text-ink-1">
									{{ selectedFromLang?.label || '' }}
								</div>
								<div class="language-label text-overline text-ink-3">
									{{ $t('bex.source_lang_label') }}
								</div>
							</div>
							<q-icon name="sym_r_expand_more" size="16px" class="text-ink-3" />
						</div>
					</div>
					<q-menu
						v-model="showFromLangMenu"
						class="bg-background-2"
						style="border-radius: 8px"
						anchor="bottom left"
						self="top left"
					>
						<q-list class="q-pa-xs">
							<q-item
								v-for="(item, index) in fromLangOptions"
								:key="index"
								:clickable="!item.disable"
								class="menu-item"
								:class="
									item.value === rule.fromLang ? 'menu-item-selected' : ''
								"
								@click="handleFromLangChange(item)"
								v-close-popup
							>
								<q-item-section
									class="text-body2 nowrap"
									:class="
										!item.disable
											? item.value === rule.fromLang
												? 'text-blue-6'
												: 'text-ink-2'
											: 'text-grey-4'
									"
								>
									<div class="ellipsis full-width">{{ item.label }}</div>
								</q-item-section>
								<q-item-section side v-show="item.value === rule.fromLang">
									<q-icon
										name="sym_r_check_circle"
										size="18px"
										class="text-blue-6"
									/>
								</q-item-section>
							</q-item>
						</q-list>
					</q-menu>
				</div>

				<!-- Arrow Icon -->
				<div class="arrow-icon">
					<q-icon name="sym_r_arrow_forward" size="16px" class="text-ink-3" />
				</div>

				<!-- Target Language Selector -->
				<div class="language-selector-wrapper">
					<div
						class="language-selector"
						:class="{ 'selector-disabled': !contentScriptReady }"
						@click.stop="
							contentScriptReady && (showToLangMenu = !showToLangMenu)
						"
					>
						<div class="language-selector-content">
							<div class="language-selector-text">
								<div class="language-name text-subtitle3 text-ink-1">
									{{ selectedToLang?.label || '' }}
								</div>
								<div class="language-label text-overline text-ink-3">
									{{ $t('bex.target_lang_label') }}
								</div>
							</div>
							<q-icon name="sym_r_expand_more" size="16px" class="text-ink-3" />
						</div>
					</div>
					<q-menu
						v-model="showToLangMenu"
						class="bg-background-2"
						style="border-radius: 8px"
						anchor="bottom left"
						self="top left"
					>
						<q-list class="q-pa-xs">
							<q-item
								v-for="(item, index) in toLangOptions"
								:key="index"
								:clickable="!item.disable"
								class="menu-item"
								:class="item.value === rule.toLang ? 'menu-item-selected' : ''"
								@click="handleToLangChange(item)"
								v-close-popup
							>
								<q-item-section
									class="text-body2 nowrap"
									:class="
										!item.disable
											? item.value === rule.toLang
												? 'text-blue-6'
												: 'text-ink-2'
											: 'text-grey-4'
									"
								>
									<div class="ellipsis full-width">{{ item.label }}</div>
								</q-item-section>
								<q-item-section side v-show="item.value === rule.toLang">
									<q-icon
										name="sym_r_check_circle"
										size="18px"
										class="text-blue-6"
									/>
								</q-item-section>
							</q-item>
						</q-list>
					</q-menu>
				</div>
			</div>

			<!-- Translation Service Selector -->
			<div>
				<div class="text-body3 text-ink-3">
					{{ $t('bex.translate_service') }}
				</div>
				<div
					class="q-mt-xs"
					:class="{ 'selector-disabled': !contentScriptReady }"
				>
					<BtSelect
						v-model="rule.translator"
						:options="options"
						:border="true"
						:height="40"
						:iconSize="24"
						classes="q-px-md"
						menuClasses="q-pa-xs"
						:menuItemHeight="40"
						:disable="!contentScriptReady"
						@update:model-value="translateHandler2"
					></BtSelect>
				</div>
			</div>

			<!-- Show Translation Only Toggle -->
			<div :class="{ 'selector-disabled': !contentScriptReady }">
				<div class="row items-center justify-between">
					<span class="text-Body3 text-ink-1">
						{{ $t('bex.trans_only') }}
					</span>
					<bt-switch
						class="custom-toggle-wrapper"
						truthy-track-color="light-blue-default"
						:model-value="rule.transOnly === 'true'"
						size="xs"
						:disable="!contentScriptReady"
						@update:model-value="handleTransOnlyChange"
					/>
				</div>
			</div>

			<!-- Always Translate This Site Toggle -->
			<div :class="{ 'selector-disabled': !contentScriptReady }">
				<div class="row items-center justify-between">
					<span class="text-Body3 text-ink-1">
						{{ $t('bex.translate_page_always') }}
					</span>
					<bt-switch
						class="custom-toggle-wrapper"
						truthy-track-color="light-blue-default"
						v-model="rule.transOpen"
						size="xs"
						:disable="!contentScriptReady"
						@update:model-value="translateHandler"
					/>
				</div>
			</div>

			<!-- Translate This Page Button -->
			<CustomButton
				color="yellow-default"
				class="full-width"
				:class="{ 'btn-disabled': !contentScriptReady }"
				:disable="!contentScriptReady"
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
							transOpen ? $t('bex.show_original') : $t('bex.translate_page')
						}}</span>
					</div>
				</template>
			</CustomButton>
		</div>

		<q-inner-loading :showing="loading"> </q-inner-loading>
	</PageCard>
</template>

<script setup lang="ts">
import PageCard from 'src/pages/Plugin/components/PageCard.vue';
import BtSelect from 'src/components/base/BtSelect.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useTranslate } from 'src/composables/mobile/useTranslate';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import {
	bexFrontBusOff,
	bexFrontBusOn
} from 'src/platform/interface/bex/utils';
import checkedIcon from 'src/assets/plugin/checked.svg';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import {
	createTabChangeListenerInCurrentWindow,
	TAB_CHANGE_TYPE
} from 'src/utils/bex/tabs';
let listener;

const {
	handleTransToggle,
	options,
	translateHandler2,
	transOpen,
	translateHandler,
	loading,
	getTransRule,
	rule,
	handleFieldChange,
	handleTransOnlyChange,
	fromLangOptions,
	toLangOptions,
	contentScriptReady,
	checkContentScriptReady
} = useTranslate();

const showFromLangMenu = ref(false);
const showToLangMenu = ref(false);

const selectedFromLang = computed(() => {
	return fromLangOptions.value.find((e) => e.value === rule.fromLang);
});

const selectedToLang = computed(() => {
	return toLangOptions.value.find((e) => e.value === rule.toLang);
});

const handleFromLangChange = (item) => {
	if (!contentScriptReady.value) {
		return;
	}
	if (!item.disable) {
		handleFieldChange('fromLang', item.value);
	}
};

const handleToLangChange = (item) => {
	if (!contentScriptReady.value) {
		return;
	}
	if (!item.disable) {
		handleFieldChange('toLang', item.value);
	}
};

onMounted(() => {
	getTransRule();
	listener = createTabChangeListenerInCurrentWindow((info) => {
		if (info.type === TAB_CHANGE_TYPE.STATUS_CHANGE) {
			return;
		}

		getTransRule();
	});

	bexFrontBusOn('TRANSLATE_SCRIPT_READY', () => {
		checkContentScriptReady();
	});
});

onUnmounted(() => {
	listener && listener.remove();
	bexFrontBusOff('TRANSLATE_SCRIPT_READY');
});
</script>

<style lang="scss" scoped>
.custom-toggle-wrapper {
	::v-deep(.q-toggle__inner--truthy .q-toggle__thumb:after) {
		background-color: $ink-on-brand !important;
	}
}

.btn-disabled {
	opacity: 0.5;
	cursor: not-allowed !important;
	pointer-events: none;
}

.selector-disabled {
	opacity: 0.5;
	cursor: not-allowed !important;
	pointer-events: none;
}

.language-selector-wrapper {
	flex: 1;
	min-width: 100px;
}

.language-selector {
	border: 1px solid $separator-2;
	border-radius: 8px;
	padding: 8px 8px 8px 12px;
	cursor: pointer;
	background: $background-1;
	transition: background 0.2s;

	&:hover {
		background: $background-3;
	}
}

.language-selector-content {
	display: flex;
	align-items: center;
	justify-content: space-between;
	gap: 8px;
}

.language-selector-text {
	flex: 1;
	min-width: 0;
	display: flex;
	flex-direction: column;
}

.language-name {
	font-size: 12px;
	font-weight: 500;
	line-height: 16px;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.language-label {
	font-size: 10px;
	font-weight: 400;
	line-height: 12px;
	white-space: nowrap;
	overflow: hidden;
	text-overflow: ellipsis;
}

.arrow-icon {
	flex-shrink: 0;
	width: 16px;
	height: 16px;
	display: flex;
	align-items: center;
	justify-content: center;
}

.menu-item {
	height: 40px;
	min-height: 40px;
	border-radius: 4px;
	padding: 8px 8px 8px 12px;
	white-space: nowrap;

	::v-deep(.q-item__section) {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
}

.menu-item-selected {
	background-color: $background-3;
}
</style>
