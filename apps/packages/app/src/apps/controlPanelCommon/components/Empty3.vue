<template>
	<div class="column items-center absolute-center">
		<q-img :src="empty3Icon" :ratio="1" width="160px" />
		<div class="text-body3 text-ink-2 q-mb-lg">
			{{ $t('NO_MATCHING_RESULT_FOUND') }}
		</div>
		<QButtonStyle v-if="refresh">
			<q-btn
				dense
				outline
				no-caps
				@click="clickHandler"
				color="blue-default"
				class="btn-wrapper"
			>
				<div class="q-px-sm">
					<q-icon name="sym_r_refresh" size="20px" color="blue-default" />
					<span class="text-body2 q-ml-xs">{{ $t('REFRESH_DATA') }}</span>
				</div>
			</q-btn>
		</QButtonStyle>
	</div>
</template>

<script setup lang="ts">
import empty3Icon from '../assets/empty3.svg';
import QButtonStyle from './QButtonStyle.vue';

interface Props {
	refresh?: boolean;
}

withDefaults(defineProps<Props>(), {
	refresh: true
});

const emits = defineEmits(['click']);

const clickHandler = () => {
	emits('click');
};
</script>

<style lang="scss" scoped>
::v-deep(.btn-wrapper.q-btn--outline:before) {
	border-color: $blue-default;
}
</style>
