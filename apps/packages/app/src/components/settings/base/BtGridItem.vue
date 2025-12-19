<template>
	<div class="bt-grid-item-root column justify-start">
		<adaptive-layout>
			<template v-slot:pc>
				<div class="row justify-start">
					<div class="text-body2 bt-grid-item-value" v-if="value">
						{{ value }}
					</div>
					<slot name="value" />
					<q-popup-proxy class="reminder_banner" v-if="popupInfo">
						<div v-html="popupInfo" />
					</q-popup-proxy>
					<q-img class="bt-grid-item-image" :src="icon" v-if="icon" />
				</div>
				<div class="text-overline bt-grid-item-label">
					{{ label }}
				</div>
			</template>
			<template v-slot:mobile>
				<template v-if="revert">
					<div class="row justify-start">
						<div class="text-body3-m bt-grid-item-value" v-if="value">
							{{ value }}
						</div>
						<slot name="value" />
						<q-popup-proxy class="reminder_banner" v-if="popupInfo">
							<div v-html="popupInfo" />
						</q-popup-proxy>
						<q-img class="bt-grid-item-image" :src="icon" v-if="icon" />
					</div>
					<div
						class="bt-grid-item-label q-mt-xs"
						:class="
							mobileTitleClasses.length ? mobileTitleClasses : 'text-overline-m'
						"
					>
						{{ label }}
					</div>
				</template>
				<template v-else>
					<div
						class="bt-grid-item-label q-mb-xs"
						:class="
							mobileTitleClasses.length ? mobileTitleClasses : 'text-overline-m'
						"
					>
						{{ label }}
					</div>
					<div class="row justify-start">
						<div class="text-body3-m bt-grid-item-value" v-if="value">
							{{ value }}
						</div>
						<slot name="value" />
						<q-popup-proxy class="reminder_banner" v-if="popupInfo">
							<div v-html="popupInfo" />
						</q-popup-proxy>
						<q-img class="bt-grid-item-image" :src="icon" v-if="icon" />
					</div>
				</template>
			</template>
		</adaptive-layout>
	</div>
</template>

<script lang="ts" setup>
import AdaptiveLayout from '../AdaptiveLayout.vue';
defineProps({
	label: {
		type: String,
		require: true
	},
	value: {
		type: String,
		required: false,
		default: ''
	},
	icon: {
		type: String,
		required: false
	},
	popupInfo: {
		type: String,
		required: false
	},
	revert: {
		type: Boolean,
		required: false
	},
	mobileTitleClasses: {
		type: String,
		default: '',
		required: false
	}
});
</script>

<style scoped lang="scss">
.bt-grid-item-root {
	.bt-grid-item-value {
		color: $ink-1;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		// width: 170px;
	}

	.bt-grid-item-image {
		margin-left: 4px;
		width: 20px;
		height: 20px;
	}

	.bt-grid-item-label {
		text-transform: capitalize;
		color: $ink-3;
	}
}
</style>
