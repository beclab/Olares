<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('Add Source')"
		@onSubmit="onOK"
		:okLoading="isLoading ? 'loading' : false"
		:cancel="t('base.cancel')"
		:okDisabled="errorMessageTitle || errorMessageURL || !url || !name"
		:ok="t('base.confirm')"
	>
		<div class="column">
			<div class="prompt-name q-mb-xs text-body3">
				{{ t('Source Title') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="name"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				no-error-icon
				placeholder=""
				@update:modelValue="validateTitle"
			/>

			<div v-if="errorMessageTitle" class="text-negative text-body3 q-mt-xs">
				{{ errorMessageTitle }}
			</div>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('Source URL') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="url"
				borderless
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				no-error-icon
				errorMessageTitle=""
				@update:modelValue="validateUrl"
			/>

			<div v-if="errorMessageURL" class="text-negative text-body3 q-mt-xs">
				{{ errorMessageURL }}
			</div>

			<div class="prompt-name q-mb-xs text-body3 q-mt-lg">
				{{ t('Description') }}
			</div>
			<q-input
				class="prompt-input text-body3"
				v-model="description"
				borderless
				no-error-icon
				input-class="text-ink-2 text-body3"
				input-style="height: 32px"
				dense
				placeholder=""
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { addMarketSource } from '../../../api/market/private/source';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useCenterStore } from '../../../stores/market/center';
import { useI18n } from 'vue-i18n';
import { ref } from 'vue';
import {
	MARKET_SOURCE_PREFIX,
	MARKET_SOURCE_TYPE
} from '../../../constant/constants';
import validator from 'validator';

const customRef = ref();
const { t } = useI18n();
const name = ref<string>('');
const url = ref<string>('');
const description = ref<string>('');
const centerStore = useCenterStore();
const isLoading = ref(false);
const errorMessageTitle = ref('');
const errorMessageURL = ref('');

const onOK = async () => {
	if (
		errorMessageTitle.value ||
		errorMessageURL.value ||
		!url.value ||
		!name.value
	) {
		return;
	}
	isLoading.value = true;

	addMarketSource({
		id: MARKET_SOURCE_PREFIX + name.value,
		name: name.value,
		base_url: url.value.trim(),
		description: description.value,
		type: MARKET_SOURCE_TYPE.REMOTE
	})
		.then((data) => {
			if (data) {
				centerStore.sources = data.sources;
				customRef.value.onDialogOK();
			}
		})
		.catch((err) => {
			notifyFailed(err.message || err.response?.data?.message || err);
		})
		.finally(() => {
			isLoading.value = false;
		});
};

const validateTitle = () => {
	if (name.value.length > 10) {
		errorMessageTitle.value = t(
			'Source Title should be less than 10 characters'
		);
		return;
	}

	const regex = /^[A-Za-z0-9]*$/;
	if (name.value && !regex.test(name.value)) {
		errorMessageTitle.value = t('Only alphanumeric characters are allowed.');
		return;
	}

	if (centerStore.sources.find((item) => item.name === name.value)) {
		errorMessageTitle.value = t(
			'Source ID already exists. Please use a different name.',
			{ sourceId: name.value }
		);
		return;
	}

	errorMessageTitle.value = '';
};

const validateUrl = () => {
	let processedUrl = url.value.trim();
	errorMessageURL.value = '';

	if (!processedUrl.startsWith('https://')) {
		errorMessageURL.value = t('httpsRequired');
		return;
	}

	if (
		!validator.isURL(processedUrl, {
			protocols: ['https'],
			require_protocol: true
		})
	) {
		errorMessageURL.value = t('invalidUrlFormat');
		return;
	}

	const exists = centerStore.sources.some((item) => {
		if (!item.base_url) return false;
		let storedUrl = item.base_url.trim();
		if (!/^https:\/\//i.test(storedUrl)) {
			storedUrl = `https://${storedUrl}`;
		}
		return storedUrl === processedUrl;
	});

	if (exists) {
		errorMessageURL.value = t(
			'Source URL already exists. Please use a different source URL'
		);
	}
};
</script>

<style scoped lang="scss">
:deep(.q-field__messages) {
	font-size: 12px;
	margin-top: 20px;
}

.prompt-name {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
}

.prompt-input {
	padding-left: 7px;
	border: 1px solid $input-stroke;
	border-radius: 8px;
	color: $ink-3;
	height: 32px;
}
</style>
