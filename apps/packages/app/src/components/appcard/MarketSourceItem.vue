<template>
	<div
		class="market-source-item row justify-between items-center"
		:class="isChanging || isDeleting ? '' : 'cursor-pointer'"
		@click="onItemClick"
	>
		<div class="row justify-end items-center">
			<q-img
				class="checkbox-image"
				:src="
					modelValue
						? 'market/circle_check_box.svg'
						: 'market/circle_uncheck_box.svg'
				"
			/>
			<div class="column justify-start q-ml-sm">
				<div class="text-body2 text-ink-1">{{ source.name }}</div>
				<div class="text-body3 text-ink-3">{{ source.base_url }}</div>
			</div>
		</div>
		<div class="row justify-end items-center">
			<div class="q-pa-xs" @click.stop>
				<q-icon size="20px" name="sym_r_info" />
				<bt-popup style="width: 243px; padding: 12px">
					<div class="text-body1 text-ink-3">{{ t('Source ID') }}</div>
					<div class="text-body1 text-ink-1 q-mt-xs">
						{{ source.id }}
					</div>
					<div class="text-body1 text-ink-3 q-mt-lg">
						{{ t('Description') }}
					</div>
					<div class="text-body1 text-ink-1 q-mt-xs">
						{{ source.description }}
					</div>
				</bt-popup>
			</div>
			<div v-if="showDeleteIcon" class="q-pa-xs" @click.stop="onItemDelete">
				<q-icon v-if="!isDeleting" size="20px" name="sym_r_delete" />
				<bt-loading v-else :loading="isDeleting" />
			</div>
		</div>
	</div>
</template>
<script setup lang="ts">
import BtLoading from '../base/BtLoading.vue';
import BtPopup from '../base/BtPopup.vue';
import { deleteMarketSource } from '../../api/market/private/source';
import { setMarketSource } from '../../api/market/private/setting';
import { notifyFailed } from '../../utils/notifyRedefinedUtil';
import { useCenterStore } from '../../stores/market/center';
import { BtDialog, useColor } from '@bytetrade/ui';
import { ref, PropType, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import {
	ALL_MARKET_OFFICIAL_SOURCES,
	MarketSource
} from '../../constant/constants';

const props = defineProps({
	modelValue: {
		type: Boolean,
		required: true
	},
	source: {
		type: Object as PropType<MarketSource>,
		required: true
	}
});

const { t } = useI18n();
const isChanging = ref(false);
const isDeleting = ref(false);
const centerStore = useCenterStore();
const { color: blue } = useColor('blue-default');
const { color: textInk } = useColor('ink-2');
const showDeleteIcon = computed(() => {
	if (props.modelValue) {
		return false;
	}
	return !ALL_MARKET_OFFICIAL_SOURCES.has(props.source.id);
});

const onItemClick = () => {
	if (isChanging.value || isDeleting.value || props.modelValue) {
		return;
	}
	BtDialog.show({
		title: t('Change Source'),
		message: t(
			'Are you sure you want to change the market sources? This might take some time to update.'
		),
		okStyle: {
			background: blue.value,
			color: textInk.value
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				isChanging.value = true;
				setMarketSource(props.source?.id)
					.then(() => {
						location.reload();
					})
					.catch((err) => {
						notifyFailed(err.message || err.response?.data?.message || err);
					})
					.finally(() => {
						isChanging.value = false;
					});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};

const onItemDelete = () => {
	const app = centerStore.getSourceInstalledApp(props.source.id);
	if (app.length > 0) {
		BtDialog.show({
			title: t('Delete Source'),
			message: t(
				'Cannot delete this source. Please uninstall all apps installed from this source, then try again.'
			),
			okStyle: {
				background: blue.value,
				color: textInk.value
			},
			okText: t('base.confirm')
		})
			.then(() => {})
			.catch((err) => {
				console.log('click error', err);
			});
		return;
	}

	BtDialog.show({
		title: t('Delete Source'),
		message: t(
			'Are you sure you want to delete this market source? This operation is permanent and cannot be undone.'
		),
		okStyle: {
			background: blue.value,
			color: textInk.value
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then((res) => {
			if (res) {
				isDeleting.value = true;
				deleteMarketSource(props.source?.id)
					.then((data) => {
						if (data) {
							centerStore.sources = data.sources;
							centerStore.removeSourceId(props.source.id);
						}
					})
					.catch((err) => {
						notifyFailed(err.message || err.response?.data?.message || err);
					})
					.finally(() => {
						isDeleting.value = false;
					});
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log('click error', err);
		});
};
</script>

<style scoped lang="scss">
.market-source-item {
	padding: 8px 12px;
	border-radius: 12px;
	border: 1px solid $separator;

	.checkbox-image {
		width: 16px;
		height: 16px;
	}
}
</style>
