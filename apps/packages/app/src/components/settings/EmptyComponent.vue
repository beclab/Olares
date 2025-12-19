<template>
	<div class="empty-parent column items-center">
		<q-img
			src="settings/default_empty.svg"
			class="empty-image"
			:style="`margin-top: ${emptyImageTop}px`"
		/>
		<div
			class="empty-text"
			:class="
				deviceStore.isMobile
					? 'text-subtitle1-m text-ink-1'
					: 'text-body2 text-ink-2'
			"
		>
			{{
				info.length > 0
					? info
					: emptyItem.length > 0
					? t('no_item_available', {
							item: emptyItem
					  })
					: ''
			}}
		</div>
		<slot name="more" />
	</div>
</template>
<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';
defineProps({
	info: {
		type: String,
		required: false,
		default: ''
	},
	emptyItem: {
		type: String,
		required: false,
		default: ''
	},
	emptyImageTop: {
		type: Number,
		default: 144,
		required: false
	}
});
const { t } = useI18n();
const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.empty-parent {
	width: 100%;
	height: calc(100% - 56px);

	.empty-image {
		width: 240px;
		height: 240px;
	}

	.empty-text {
		text-align: center;
	}
}
</style>
