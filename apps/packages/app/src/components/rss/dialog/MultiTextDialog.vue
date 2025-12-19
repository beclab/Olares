<template>
	<bt-custom-dialog
		:title="t(title)"
		ref="customRef"
		size="medium"
		@onSubmit="onOK"
		:okLoading="isLoading"
		:cancel="t('base.cancel')"
		:ok="t('base.confirm')"
		:ok-disabled="okDisabled"
	>
		<div class="row justify-start items-center q-mt-xs q-mb-sm" v-if="!link">
			<template v-for="item in options" :key="item.value">
				<bt-check-box-component
					:label="item.label"
					:circle="true"
					class="q-mr-lg"
					:model-value="modelValue === item.value"
					@update:model-value="updateModelValue(item.value)"
				/>
			</template>
		</div>
		<div
			v-if="!link && modelValue === options[options.length - 1].value"
			class="text-body1 text-ink-3 q-mt-xs"
		>
			{{ t('Domain') }}
		</div>
		<q-input
			v-if="!link && modelValue === options[options.length - 1].value"
			dense
			borderless
			v-model="domain"
			class="text-domain q-mt-xs"
			style="height: 40px"
		/>
		<div class="text-body1 text-ink-3 q-mt-xs">
			{{
				t(label, {
					number: link ? getLinksLength() : getCookiesLength()
				})
			}}
		</div>
		<q-input
			dense
			borderless
			v-model="input"
			class="text-input q-mt-xs"
			style="
				height: 120px;
				overflow: scroll;
				scrollbar-width: none;
				padding-top: 0;
			"
			:input-style="{ resize: 'none' }"
			type="textarea"
		/>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import BtCheckBoxComponent from 'src/components/settings/base/BtCheckBoxComponent.vue';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { batchEntries } from 'src/api/wise';
import { useRssStore } from 'src/stores/rss';
import { useCookieStore } from 'src/stores/settings/cookie';
import { notifyFailed, notifySuccess } from 'src/utils/settings/btNotify';

const props = defineProps({
	title: String,
	label: String,
	link: {
		type: Boolean,
		default: true
	}
});
const { t } = useI18n();
const isLoading = ref(false);
const customRef = ref();
const domain = ref();
const input = ref();
const modelValue = ref(0);
const okDisabled = computed(
	() => !input.value && cookieStore.getCookieCount(input.value).length > 0
);
const options = [
	{
		label: 'Netscape',
		value: 0
	},
	{
		label: 'JSON',
		value: 1
	},
	{
		label: 'Header String',
		value: 2
	}
];
const rssStore = useRssStore();
const cookieStore = useCookieStore();
const updateModelValue = (item: number) => {
	modelValue.value = item;
};

const onOK = async () => {
	if (props.link) {
		isLoading.value = true;
		batchEntries(input.value)
			.then(() => {
				notifySuccess('success');
				customRef.value.onDialogOK();
				rssStore.sync();
			})
			.catch((err) => {
				notifyFailed(err.message);
			})
			.finally(() => {
				isLoading.value = false;
			});
	} else {
		switch (modelValue.value) {
			case 0:
				isLoading.value = true;
				cookieStore
					.addNetscapeCookies(input.value)
					.then(() => {
						notifySuccess('success');
						customRef.value.onDialogOK();
					})
					.catch((err) => {
						notifyFailed(err.message);
					})
					.finally(() => {
						isLoading.value = false;
					});
				break;
			case 1:
				isLoading.value = true;
				cookieStore
					.addJsonCookies(input.value)
					.then(() => {
						notifySuccess('success');
						customRef.value.onDialogOK();
					})
					.catch((err) => {
						notifyFailed(err.message);
					})
					.finally(() => {
						isLoading.value = false;
					});
				break;
			case 2:
				isLoading.value = true;
				cookieStore
					.addHeaderCookies(input.value, domain.value)
					.then(() => {
						notifySuccess('success');
						customRef.value.onDialogOK();
					})
					.catch((err) => {
						notifyFailed(err.message);
					})
					.finally(() => {
						isLoading.value = false;
					});
				break;
			default:
				return;
		}
	}
};

function getLinksLength(): number {
	if (!input.value) return 0;
	const lines = input.value.split(/\r?\n/);
	const nonEmptyLines = lines
		.filter((line) => line.trim() !== '')
		.filter((line) => !line.startsWith('#'));
	console.log('Links ：', nonEmptyLines);
	return nonEmptyLines.length;
}

function getCookiesLength(): number {
	switch (modelValue.value) {
		case 0:
			return cookieStore.getNetscapeCookieCount(input.value);
		case 1:
			return cookieStore.getJsonCookieCount(input.value);
		case 2:
			return cookieStore.getHeaderCookieCount(input.value, domain.value);
		default:
			return 0;
	}
}
</script>
<style scoped lang="scss">
.text-domain {
	border-radius: 8px;
	border: 1px solid $input-stroke;
	padding: 0 12px;
}

.text-input {
	border-radius: 8px;
	border: 1px solid $input-stroke;
	padding: 10px 12px;
}

::v-deep(.q-textarea .q-field__native) {
	padding-top: 10px;
}
</style>
