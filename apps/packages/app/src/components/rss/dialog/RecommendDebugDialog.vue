<template>
	<bt-custom-dialog
		ref="customRef"
		size="medium"
		:title="t('base.debug_info')"
		@onSubmit="onCopy"
		:ok="t('base.copy')"
		:cancel="t('base.close')"
	>
		<bt-scroll-area class="debug-scroll-area">
			<div class="debug-info">
				{{ data }}
			</div>
		</bt-scroll-area>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { copyToClipboard } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { ref } from 'vue';

const { t } = useI18n();
const customRef = ref();

const props = defineProps({
	data: { type: Object }
});

const onCopy = () => {
	if (props.data) {
		copyToClipboard(JSON.stringify(props.data))
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('copy_success')
				});
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('copy_failure_message', e.message)
				});
			});
	}
};
</script>

<style scoped lang="scss">
::v-deep(.q-dialog__inner--minimized > div) {
	max-width: calc(100vw - 200px);
	min-width: 70vw;
}
.debug-scroll-area {
	width: 100%;
	height: 400px;
	border-radius: 4px;
	color: #b7c4d1;
	background: #242e42;

	.debug-info {
		white-space: pre-wrap;
		padding: 20px;
	}
}
</style>
