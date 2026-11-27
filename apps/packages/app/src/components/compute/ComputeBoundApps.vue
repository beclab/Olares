<template>
	<div class="compute-bound-apps" :class="bgColor">
		<!-- header: label + (collapsed avatar stack) + toggle -->
		<div
			class="compute-bound-apps__header row items-center justify-between"
			:class="{ 'cursor-pointer': collapsible }"
			@click="toggle"
		>
			<span class="text-body3 text-ink-3">{{ t('compute.apps_bound') }}</span>
			<div v-if="collapsible" class="row items-center no-wrap">
				<app-avatar-stack
					v-if="!expanded"
					:apps="apps"
					:max="5"
					:show-status-dot="false"
				/>
				<q-icon
					:name="
						expanded ? 'sym_r_keyboard_arrow_up' : 'sym_r_keyboard_arrow_down'
					"
					size="20px"
					class="text-ink-3 q-ml-sm"
				/>
			</div>
		</div>

		<!-- expanded: per-app rows with unbind -->
		<q-slide-transition>
			<div v-if="isExpanded" class="compute-bound-apps__list">
				<div
					v-for="app in apps"
					:key="app.appName"
					class="compute-bound-apps__row row items-center justify-between no-wrap"
				>
					<div class="compute-bound-apps__info col row items-center no-wrap">
						<q-img
							no-spinner
							:src="app.icon"
							class="compute-bound-apps__icon"
						/>
						<div class="column q-ml-sm compute-bound-apps__meta">
							<div
								class="row justify-start items-center no-wrap"
								style="gap: 8px"
							>
								<span class="text-body2 text-ink-1 ellipsis">
									{{ app.title }}
								</span>
								<span class="text-info text-body3" v-if="app.memory > 0">{{
									formatComputeMemory(app.memory, '')
								}}</span>
							</div>
							<div
								v-if="showBadges && appBadges(app).length > 0"
								class="row items-center"
								style="gap: 4px"
							>
								<compute-resource-badge
									v-for="badge in appBadges(app)"
									:key="badge.label"
									:label="badge.i18nKey ? t(badge.label) : badge.label"
									tone="neutral"
								/>
							</div>
						</div>
					</div>

					<bt-loading
						v-if="busyAppName === app.appName"
						class="compute-bound-apps__action q-ml-sm"
						:loading="true"
						size="20px"
					/>
					<div
						v-else-if="unbindable"
						class="compute-bound-apps__unbind compute-bound-apps__action text-body3 text-ink-2 q-ml-sm"
						@click="$emit('unbind', app)"
					>
						{{ t('compute.unbind') }}
					</div>
				</div>
			</div>
		</q-slide-transition>
	</div>
</template>

<script setup lang="ts">
import BtLoading from 'src/components/base/BtLoading.vue';
import ComputeResourceBadge from 'src/components/compute/ComputeResourceBadge.vue';
import AppAvatarStack from 'src/components/compute/AppAvatarStack.vue';
import {
	AppComputeBadge,
	appComputeBadges,
	formatComputeMemory
} from 'src/constant/compute';
import { AcceleratorAppView } from 'src/stores/settings/accelerator';
import { computed, PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	apps: {
		type: Array as PropType<AcceleratorAppView[]>,
		default: () => []
	},
	/** appName currently being unbound — shows a spinner in place of the button. */
	busyAppName: {
		type: String,
		default: ''
	},
	/** Whether per-app unbind buttons are shown. */
	unbindable: {
		type: Boolean,
		default: true
	},
	collapsible: {
		type: Boolean,
		default: true
	},
	defaultExpanded: {
		type: Boolean,
		default: false
	},
	showBadges: {
		type: Boolean,
		default: true
	},
	bgColor: {
		type: String,
		default: 'bg-background-6'
	}
});

defineEmits<{ (e: 'unbind', app: AcceleratorAppView): void }>();

const { t } = useI18n();
const expanded = ref(props.defaultExpanded);
// Non-collapsible callers always render the full list.
const isExpanded = computed(() => !props.collapsible || expanded.value);
const toggle = () => {
	if (props.collapsible) expanded.value = !expanded.value;
};

const appBadges = (app: AcceleratorAppView): AppComputeBadge[] =>
	appComputeBadges(app.mode, app.scope);
</script>

<style scoped lang="scss">
.compute-bound-apps {
	border-radius: 8px;

	&__header {
		min-height: 28px;
	}

	&__list {
		margin-top: 4px;
	}

	&__row {
		padding: 8px 0 0;
	}

	// Let the app info shrink (and ellipsis the title) instead of wrapping the
	// unbind button onto the next line when the title / allocation text is long.
	&__info {
		flex: 1 1 auto;
		min-width: 0;
	}

	&__meta {
		min-width: 0;
	}

	&__action {
		flex: 0 0 auto;
	}

	&__icon {
		width: 32px;
		height: 32px;
		border-radius: 8px;
	}

	&__unbind {
		min-height: 32px;
		cursor: pointer;
		display: inline-flex;
		align-items: center;
		border: 1px solid $btn-stroke;
		border-radius: 8px;
		padding: 8px 12px;
	}
}
</style>
