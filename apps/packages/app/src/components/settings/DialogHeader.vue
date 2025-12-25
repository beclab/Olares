<template>
	<AdaptiveLayout>
		<template v-slot:pc>
			<div class="dialog-common">
				<div class="row justify-between items-center q-pl-md q-pr-sm">
					<span class="text-ink-1 text-subtitle3 title"> {{ title }}</span>
					<q-btn
						flat
						@click="close"
						style=""
						class="q-ml-sm btn-size-sm btn-no-text btn-no-border text-ink-3"
						icon="sym_r_close"
					>
						<bt-tooltip :label="t('close')"
					/></q-btn>
				</div>
			</div>
		</template>
		<template v-slot:mobile>
			<div
				class="dialog-common-mobile row items-end justify-center text-subtitle2-m text-ink-1"
			>
				{{ title }}
			</div>
		</template>
	</AdaptiveLayout>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import AdaptiveLayout from './AdaptiveLayout.vue';
import BtTooltip from './base/BtTooltip.vue';
defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	}
});
const close = () => {
	emit('closeAction');
};

const emit = defineEmits(['closeAction']);
const { t } = useI18n();
</script>

<style scoped lang="scss">
.dialog-common {
	width: 100%;
	padding: 0;
	background-color: $background-3;

	.title {
		height: 100%;
	}
}

.dialog-common-mobile {
	height: 44px;
}
</style>
