<template>
	<div class="row items-center justify-between title-part">
		<div class="text-ink-1 text-subtitle2-m">{{ title }}</div>
		<div v-if="showExtend" class="row items-center" @click="expandAction">
			<span class="text-ink-2 text-caption-m">{{
				expand ? t('files.Collapse') : t('files.Expand')
			}}</span>
			<q-icon
				:name="expand ? 'sym_r_keyboard_arrow_up' : 'sym_r_keyboard_arrow_down'"
				color="ink-2"
				size="24px"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
const props = defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	expand: {
		type: Boolean,
		required: false,
		default: false
	},
	showExtend: {
		type: Boolean,
		required: false,
		default: false
	}
});

const { t } = useI18n();

const emit = defineEmits(['revertExpand']);

const expandAction = () => {
	emit('revertExpand');
};
</script>

<style scoped lang="scss">
.title-part {
	height: 40px;
	width: 100%;

	.reset {
		cursor: pointer;
	}
}
</style>
