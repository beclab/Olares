<template>
	<div class="q-pa-lg bex-home-container">
		<div class="column no-wrap flex-gap-y-lg">
			<div>
				<span class="text-h4 text-ink-1">{{ $t('bex.welcome') }} 🌟</span>
			</div>
			<template v-if="appsStore.allAppIds.length > 0">
				<AccountHeader />
				<div>
					<span class="text-h5 text-ink-1">{{ $t('bex.quick_access') }}</span>
					<div class="q-mt-sm apps-list">
						<ForegroundApplications2></ForegroundApplications2>
					</div>
				</div>
				<div v-if="actionShow">
					<div class="text-h6 text-ink-1">{{ $t('bex.current_page') }}</div>
					<div
						class="q-pa-sm bg-background-6 row no-wrap items-center flex-gap-sm q-mt-sm current-page-wrapper"
					>
						<div class="avatar-container">
							<q-img
								:src="item?.image || mask_group"
								:error-src="mask_group"
								width="60px"
								height="60px"
								spinner-size="32px"
								crossorigin="anonymous"
								referrerpolicy="no-referrer"
								class="bg-background-6"
							>
								<template #loading>
									<q-skeleton
										type="rect"
										square
										width="60px"
										height="60px"
										animation="fade"
									/>
								</template>
							</q-img>
						</div>
						<div
							class="column flex-gap-y-xs no-wrap"
							style="flex: 1; overflow: hidden"
						>
							<div class="text-body2 text-ink-1 ellipsis-2-lines">
								{{ item?.title }}
							</div>
							<div class="text-body3 text-ink-3 ellipsis">
								{{ item?.url }}
							</div>
						</div>
					</div>
					<div class="row items-center no-wrap flex-gap-sm q-mt-sm">
						<WiseAbilityTooltipContainer
							:tooltip="collectTooltip"
							v-if="wiseActionShow"
						>
							<CustomButton
								outline
								class="q-px-md"
								@click="
									() =>
										item.status === RssStatus.none
											? onSaveEntry(item)
											: openWise()
								"
								:loading="appAbilitiesStore.loading || collectStore.loading"
								:disable="
									(!!missingAbility || !!collectTooltip) &&
									item.status === RssStatus.none
								"
							>
								<template #label>
									<span class="row items-center no-wrap flex-gap-xs">
										<q-icon
											color="ink-1"
											:name="
												item.status === RssStatus.none ||
												!appAbilitiesStore.wise.running
													? 'sym_r_box_add'
													: 'sym_r_open_in_new'
											"
											size="16px"
										/>
										<span class="text-ink-1 text-subtitle3 ellipsis">{{
											item.status === RssStatus.added &&
											appAbilitiesStore.wise.running
												? $t('bex.open_in_wise')
												: $t('collect')
										}}</span>
									</span>
								</template>
							</CustomButton>
						</WiseAbilityTooltipContainer>
						<CustomButton
							outline
							class="q-px-md"
							@click="handleTransToggle"
							:loading="translateLoading"
						>
							<template #label>
								<div
									class="row items-center flex-gap-xs no-wrap"
									style="white-space: nowrap"
								>
									<div
										class="relative-position"
										style="height: 16px; height: 16px; flex: 0 0 16px"
									>
										<q-icon
											name="sym_r_translate"
											size="16px"
											color="ink-1"
											class="absolute-center"
										/>
										<img
											:src="checkedIcon"
											alt="checked"
											style="width: 8px; height: 8px"
											class="absolute-bottom-right z-top"
											v-show="transOpen"
										/>
									</div>
									<span class="text-ink-1 text-subtitle3">{{
										transOpen ? $t('bex.show_original') : $t('bex.translate')
									}}</span>
								</div>
							</template>
						</CustomButton>
						<CookieUploadButton v-if="wiseActionShow"></CookieUploadButton>
					</div>
				</div>
			</template>
			<EmptyData
				v-else
				title="Oops! Connection Lost"
				subtitle="Check your network or try again later."
				class="absolute-center"
				@click="retryHandler"
			></EmptyData>
			<q-inner-loading :showing="appsStore.loading"> </q-inner-loading>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useUserStore } from 'src/stores/user';
import ForegroundApplications2 from 'src/pages/Plugin/containers/ForegroundApplications2.vue';
import { useAppsStore } from 'src/stores/bex/apps';
import { computed, onMounted, onUnmounted, watch } from 'vue';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { createTabChangeListenerInCurrentWindow } from 'src/utils/bex/tabs';
import AccountHeader from 'src/pages/Plugin/containers/AccountHeader.vue';
import { useTranslate } from 'src/composables/mobile/useTranslate';
import checkedIcon from 'src/assets/plugin/checked.svg';
import { useCollect } from 'src/composables/bex/useCollect';
import WiseAbilityTooltipContainer from 'src/components/WiseAbilityTooltipContainer.vue';
import mask_group from 'src/assets/common/mask_group.svg';
import { URL_VALID_STATUS } from 'src/utils/url2';
import CookieUploadButton from 'src/pages/Mobile/collect/CookieContent2.vue';
import { useCookieStatus } from 'src/composables/bex/useCookieStatus';
import { useWiseAbility } from 'src/composables/common/useWiseAbility';
let listener;

const appsStore = useAppsStore();
const userStore = useUserStore();
const { missingAbility } = useWiseAbility();
const {
	collectStore,
	RssStatus,
	onSaveEntry,
	openWise,
	item,
	init: collectInit,
	appAbilitiesStore,
	handleActivated,
	handleUpdated,
	validate
} = useCollect();

const {
	handleTransToggle,
	transOpen,
	loading: translateLoading,
	getTransRule
} = useTranslate();

const { ytdlpRequire, cookieRequire, cookieIcon, collectSiteStore } =
	useCookieStatus();

const collectTooltip = computed(() => {
	let tooltip = '';
	if (!validate.value.valid && validate.value?.reason) {
		tooltip = validate.value.reason;
	} else if (cookieRequire.value && cookieIcon.value.tooltip) {
		tooltip = cookieIcon.value.tooltip;
	} else if (!!collectSiteStore.data.cookie.is_entry_available) {
		tooltip = collectSiteStore.data.cookie.is_entry_available;
	}
	return tooltip;
});

const retryHandler = () => {
	appsStore.init();
};

const actionShow = computed(
	() =>
		validate.value.valid || validate.value.status === URL_VALID_STATUS.BLOCKED
);

const wiseActionShow = computed(
	() =>
		userStore.current_user?.isLargeVersion12 ||
		validate.value.status === URL_VALID_STATUS.BLOCKED
);

const init = () => {
	getTransRule();
	collectInit();
};

watch(
	() => appAbilitiesStore?.wise.running,
	(newValue) => {
		if (newValue) {
			init();
		}
	}
);

onMounted(() => {
	appsStore.init();
	init();

	listener = createTabChangeListenerInCurrentWindow((info) => {
		handleActivated(info);
	});
});

onUnmounted(() => {
	listener && listener.remove();
});
</script>

<style lang="scss" scoped>
.bex-home-container {
	.apps-list {
		border-radius: 12px;
		border: 1px solid $separator-2;
	}
	.current-page-wrapper {
		border-radius: 12px;
		.avatar-container {
			flex: 0 0 60px;
			border-radius: 12px;
			overflow: hidden;
		}
	}
}
.decoration-line {
	text-decoration-line: underline;
	text-decoration-style: solid;
	text-decoration-skip-ink: none;
	text-decoration-thickness: auto;
	text-underline-offset: auto;
	text-underline-position: from-font;
	cursor: pointer;
}
</style>
