<template>
	<div class="text-plus" @click="handleCopy">
		<span :class="{ link: !!to }" @click="handleClick"> {{ text }}</span>
		<div v-if="copy && text" class="copy-ico">
			<q-icon
				class="text-plus-icon cursor-pointer"
				name="sym_r_content_copy"
				size="16px"
				color="ink-3"
			/>
		</div>
	</div>
</template>

<script setup>
import { useRouter } from 'vue-router';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';

const router = useRouter();
const $q = useQuasar();
const { t } = useI18n();

const props = defineProps({
	text: String,
	copy: Boolean,
	to: null
});

const handleCopy = () => {
	const textarea = document.createElement('textarea');
	textarea.value = props.text;
	document.body.appendChild(textarea);
	textarea.select();
	document.execCommand('copy');
	document.body.removeChild(textarea);

	$q.notify({
		type: 'positive',
		message: t('COPY_SUCCESSFUL'),
		position: 'top'
	});
};

const handleClick = () => {
	if (props.to) {
		router.push(props.to);
	}
};
</script>

<style lang="scss">
.text-plus {
	display: inline-flex;
	gap: 4px;
	.link {
		color: var(--el-color-primary);
		cursor: pointer;
		&:hover {
			opacity: 0.8;
		}
	}
}
</style>
