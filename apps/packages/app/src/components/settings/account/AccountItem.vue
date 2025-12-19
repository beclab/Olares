<template>
	<q-item
		:class="{
			'q-list-class': border && !deviceStore.isMobile,
			'mobile-items-list': border && deviceStore.isMobile,
			'q-pa-none': !border,
			'account-item': border && !deviceStore.isMobile
		}"
		:style="border ? 'height: 72px' : 'padding: 16px 20px'"
		:clickable="clickable"
		class="text-background-1"
		@click="accountClick"
	>
		<q-item-section>
			<div class="row items-center">
				<q-img
					width="40px"
					height="40px"
					:noSpinner="true"
					:src="`account/${imgName}.png`"
					v-if="imgName && imgName.length"
				/>
				<slot name="avatar" v-else />
				<div
					class="column justify-start info-part q-ml-sm"
					:class="detail && detail.length > 0 ? '' : 'justify-center'"
				>
					<div class="row">
						<div
							class="text-ink-1 content"
							:class="{
								'text-subtitle2': !deviceStore.isMobile,
								'text-subtitle3-m': deviceStore.isMobile
							}"
						>
							{{ title }}
						</div>
						<div
							class="status-common text-caption q-ml-sm row items-center justify-center"
							:class="available ? 'status-available' : 'status-unabled'"
						>
							{{ available ? t('active') : t('inactive') }}
						</div>
					</div>
					<div
						class="account-detail q-mt-xs single-line"
						:class="{
							'text-body3': !deviceStore.isMobile,
							'text-overline-m': deviceStore.isMobile
						}"
						v-if="detail && detail.length"
					>
						{{ detail }}
					</div>
				</div>
			</div>
		</q-item-section>
		<q-item-section side v-if="side">
			<bt-check-box-component v-if="selectable" :model-value="selected" />
			<q-icon v-else name="sym_r_chevron_right" color="ink-1" size="24px" />
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';
import BtCheckBoxComponent from '../base/BtCheckBoxComponent.vue';

defineProps({
	imgName: {
		required: false,
		default: '',
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
	},
	side: {
		required: false,
		default: true,
		type: Boolean
	},
	clickable: {
		required: false,
		default: true,
		type: Boolean
	},
	available: {
		required: true,
		type: Boolean
	},
	border: {
		type: Boolean,
		required: false,
		default: true
	},
	selectable: {
		type: Boolean,
		required: false,
		default: false
	},
	selected: {
		type: Boolean,
		required: false,
		default: false
	}
});

const { t } = useI18n();
const deviceStore = useDeviceStore();

const accountClick = () => {
	emit('accountClick');
};

const emit = defineEmits(['accountClick']);
</script>

<style scoped lang="scss">
.status-common {
	height: 20px;
	border-radius: 4px;
	padding: 4px 8px;
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
	max-width: calc(100% - 20px);
	color: $ink-3;
}
.info-part {
	width: calc(100% - 48px);
	.content {
		max-width: calc(100% - 60px);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}

.account-item:hover {
	background-color: $background-hover;
}
</style>
