<template>
	<div class="mobile-title text-subtitle1 text-ink-1" v-if="isMobile">
		{{ label }}
	</div>

	<div class="dialog-header row items-center justify-between" v-else>
		<q-icon
			v-if="icon"
			color="ink-1"
			:name="icon"
			size="18px"
			class="q-mr-sm"
		/>
		<div
			class="title text-subtitle1 col text-ink-1"
			:class="titAlign ? titAlign : 'text-left'"
		>
			{{ label }}
		</div>
		<q-space />
		<q-btn dense flat icon="close" size="sm" color="ink-3" @click="onCancel">
			<q-tooltip>{{ t('buttons.close') }}</q-tooltip>
		</q-btn>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import { isPad } from '../../utils/platform';

defineProps({
	label: {
		type: String,
		default: '',
		required: false
	},
	icon: {
		type: String,
		default: '',
		required: false
	},
	titAlign: {
		type: String,
		default: 'left',
		required: false
	}
});

const $q = useQuasar();
const isMobile = ref(
	(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile) && !isPad()
);
const emit = defineEmits(['close']);
const { t } = useI18n();

const onCancel = () => {
	emit('close');
};
</script>

<style scoped lang="scss">
.mobile-title {
	color: $ink-1;
	text-align: center;
	margin: 20px 0;
}

.dialog-header {
	height: 56px;
	line-height: 56px;
	padding: 0 20px;
}
</style>
