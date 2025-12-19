<template>
	<q-item :clickable="false" class="analysis-item q-pa-none" dense>
		<q-item-section>
			<div class="row items-center" style="width: 100%">
				<q-img
					width="32px"
					height="32px"
					:noSpinner="true"
					:src="getRequireImage(imgName)"
				/>
				<div
					class="column justify-start info-part"
					:class="detail && detail.length > 0 ? '' : 'justify-center'"
					style="margin-left: 8px"
				>
					<div
						class="row"
						:style="{
							'max-width': '100%'
						}"
					>
						<div class="text-subtitle2 account-title">
							{{ title }}
						</div>
					</div>
					<div v-if="$slots.title">
						<slot name="title" />
					</div>
					<div
						class="text-body3 account-detail q-mt-xs"
						v-if="detail && detail.length"
					>
						{{ detail }}
					</div>
				</div>
			</div>
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
// import { useI18n } from 'vue-i18n';
import { getRequireImage } from '../../../utils/imageUtils';
defineProps({
	imgName: {
		required: false,
		default: 'cloud/folder.svg',
		type: String
	},

	title: {
		required: false,
		default: '',
		type: String
	},
	detail: {
		required: false,
		default: '',
		type: String
	}
});

// const { t } = useI18n();
</script>

<style scoped lang="scss">
.status-common {
	height: 20px;
	border-radius: 4px;
}

.status-available {
	background: $green-soft;
	color: $positive;
}

.status-unabled {
	background: $red-soft;
	color: $negative;
}

.account-detail {
	color: $ink-3;
	overflow: hidden;
	text-overflow: ellipsis;
	// max-width: 320px;
}

.border-item {
	border: 1px solid $separator;
	border-radius: 12px;
	padding: 0 16px;
	margin-top: 12px;
}

.analysis-item {
	// height: 36px;
}

.info-part {
	width: calc(100% - 48px);
	.account-title {
		color: $ink-1;
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		word-break: break-word;
	}
}
</style>
