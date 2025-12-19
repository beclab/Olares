<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="$t('dialog.add_link')"
		@show="onDialogShow"
		:ok="false"
	>
		<div class="relative-position">
			<div class="column no-wrap flex-gap-y-lg">
				<div>
					<div class="text-body2 text-ink-3">
						{{ link ? $t('sharedUrl') : $t('inputUrl') }}
					</div>
					<div class="q-mt-xs">
						<q-input
							ref="inputRef"
							class="prompt-input text-body3"
							v-model="url"
							borderless
							input-class="text-ink-2 text-body3"
							input-style="height: 42px"
							dense
							:debounce="1000"
							@update:model-value="onClick"
							:loading="collectSiteStore.loading"
							:disable="!!link"
						>
							<template #loading>
								<SpinnerLoading></SpinnerLoading>
							</template>
						</q-input>
					</div>
					<div
						v-if="!validate.valid && url"
						class="text-body3 text-negative q-mt-xs"
					>
						{{ validate.reason }}
					</div>
				</div>
				<collection-content />
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, onMounted, provide, ref } from 'vue';
import { useCollectSiteStore } from 'src/stores/collect-site';
import CollectionContent from 'src/pages/Plugin/collect/CollectionContent.vue';
import {
	UrlValidationResult,
	validateUrlWithReasonAsync
} from 'src/utils/url2';
import SpinnerLoading from 'src/components/common/SpinnerLoading.vue';
import { COLLECT_THEME } from 'src/constant/provide';
import { WISE_COLLECT_THEME } from 'src/constant/theme';
import { useTerminusStore } from 'src/stores/terminus';
import { useConfigStore } from 'src/stores/rss-config';
import { useBrowserCookieStore } from 'src/stores/settings/browserCookie';

provide(COLLECT_THEME, WISE_COLLECT_THEME);

interface Props {
	link?: string;
}

const props = withDefaults(defineProps<Props>(), {});

const validateDefault = { valid: false };
const collectSiteStore = useCollectSiteStore();
collectSiteStore.init();
const url = ref(props.link || '');
const inputRef = ref();
const validate = ref<UrlValidationResult>({ ...validateDefault });
const browserCookieStore = useBrowserCookieStore();

const onClick = async () => {
	browserCookieStore.current_tab = undefined;
	validate.value = await validateUrlWithReasonAsync(url.value);
	if (!validate.value.valid || !url.value) {
		collectSiteStore.reset();
		return;
	}
	validReset();
	collectSiteStore.search(url.value);
	getCookie(url.value);
};

const getCookie = (urlTarget: string) => {
	const terminusStore = useTerminusStore();
	const olaresId = terminusStore.olaresId.split('@')[0];
	const rssStore = useConfigStore();

	const url =
		process.env.NODE_ENV === 'development'
			? ''
			: rssStore.getModuleSever('settings');

	browserCookieStore.init(
		{
			url: urlTarget
		},
		olaresId,
		url
	);
};

onMounted(() => {
	if (props.link) {
		onClick();
	}
});

const validReset = () => {
	validate.value = { ...validateDefault };
};
onBeforeUnmount(() => {
	collectSiteStore.reset();
});

const onDialogShow = () => {
	inputRef.value?.focus();
};
</script>

<style scoped lang="scss">
.feed-icon {
	width: 32px;
	height: 32px;
}

.loading-item {
	height: 40px;
}

.subscribe-item {
	margin-bottom: 8px;
}

.subscribe-item:last-child {
	margin-bottom: 0;
}

.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.prompt-input {
	padding-left: 12px;
	padding-right: 12px;
	height: 42px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
	border: 1px solid $input-stroke;
}
.scroll-wrapper {
	::v-deep(.q-scrollarea__content) {
		width: 100%;
	}
}
</style>
