<template>
	<div class="preview-share-container row items-center justify-between q-pa-md">
		<div
			class="preview-share-edit row justify-between q-px-md q-py-sm"
			:style="{
				width: userStore.isMobile ? 'calc(100% - 110px)' : 'calc(100% - 92px)',
				height: userStore.isMobile ? '40px' : '32px'
			}"
		>
			<div
				:class="userStore.isMobile ? 'text-body1' : 'text-body3'"
				class="preview-share-url text-ink-2"
				:style="{
					width: userStore.isMobile ? 'calc(100% - 60px)' : 'calc(100% - 44px)'
				}"
			>
				{{ url }}
			</div>
			<div class="row justify-end items-center">
				<q-icon
					name="sym_r_content_copy"
					color="ink-3"
					class="cursor-pointer"
					:size="userStore.isMobile ? '24px' : '16px'"
					@click="copyUrl"
				/>
				<q-icon
					name="sym_r_share_windows"
					color="ink-3"
					class="q-ml-xs cursor-pointer"
					:size="userStore.isMobile ? '24px' : '16px'"
					@click="open"
				/>
			</div>
		</div>
		<bio-button
			:class="userStore.isMobile ? 'text-body1' : 'text-body3'"
			class="q-ml-md"
			:label="t('base.share')"
			icon="sym_r_share"
		>
			<bio-share-popup :url="url" :name="name" />
		</bio-button>
	</div>
</template>

<script lang="ts" setup>
import BioSharePopup from '@apps/profile/src/components/profile/share/BioSharePopup.vue';
import BioButton from '@apps/profile/src/components/profile/base/BioButton.vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { OlaresInfo, TerminusInfo } from '@bytetrade/core';
import { useI18n } from 'vue-i18n';
import { watch, ref } from 'vue';
import { getApplication } from 'src/application/base';

const { t } = useI18n();
const url = ref();
const zone = ref<string>('');
const name = ref<string>('');
const userStore = useUserStore();

const copyUrl = () => {
	getApplication()
		.copyToClipboard(url.value)
		.then(() => {
			BtNotify.show({
				type: NotifyDefinedType.SUCCESS,
				message: 'Copied URL to clipboard'
			});
		})
		.catch((e) => {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: 'Copied failure ' + e.message
			});
		});
};

const open = () => {
	window.open(url.value);
};

watch(
	() => userStore.info,
	(value: OlaresInfo | null) => {
		console.log(value);

		if (value) {
			zone.value =
				value.olaresId || value.terminusName
					? (value.olaresId || value.terminusName).replace('@', '.')
					: '';
			name.value =
				value.olaresId || value.terminusName
					? (value.olaresId || value.terminusName).split('@')[0]
					: '';
			url.value = `https://${zone.value}`;
		}
	},
	{
		immediate: true
	}
);
</script>
<style lang="scss">
.preview-share-container {
	border-top: var(--share-header-border);
	height: var(--share-header-height);
	width: 100%;

	.preview-share-edit {
		min-width: 150px;
		border-radius: 8px;
		border: 1px solid $input-stroke;
		overflow: hidden;

		.preview-share-url {
			white-space: nowrap;
			overflow: hidden;
			text-overflow: ellipsis;
		}
	}
}
</style>
