<template>
	<q-item
		:class="border ? 'border-item' : 'q-px-none'"
		:clickable="clickable"
		class="integration-item"
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
					class="column justify-start info-part"
					:class="detail && detail.length > 0 ? '' : 'justify-center'"
					style="margin-left: 8px"
				>
					<div class="row" style="max-width: 100%">
						<div class="text-subtitle2 account-title">
							{{ title }}
						</div>
						<div
							class="status-common text-caption q-px-md q-ml-sm row items-center justify-center"
							:class="available ? 'status-available' : 'status-unabled'"
						>
							{{ available ? t('active') : t('inactive') }}
						</div>
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
		<q-item-section side v-if="side">
			<q-icon name="sym_r_chevron_right" color="ink-1" size="20px" />
		</q-item-section>
		<q-item-section side v-else>
			<slot name="side" />
		</q-item-section>
	</q-item>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
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
	}
});

const { t } = useI18n();

const accountClick = () => {
	emit('accountClick');
};

const emit = defineEmits(['accountClick']);
</script>

<style scoped lang="scss">
.status-common {
	height: 20px;
	border-radius: 4px;
	overflow: hidden;
	text-overflow: ellipsis;
	white-space: nowrap;
	max-width: 60px;
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
}

.border-item {
	border: 1px solid $separator;
	border-radius: 12px;
	padding: 0 16px;
	margin-top: 12px;
}

.integration-item {
	height: 64px;
}

.info-part {
	width: calc(100% - 48px);
	.account-title {
		color: $ink-1;
		max-width: calc(100% - 70px);
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
}
</style>
