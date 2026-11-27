<template>
	<bt-custom-dialog
		ref="customRef"
		:title="title"
		:skip="false"
		:ok="okText || t('confirm')"
		:cancel="t('cancel')"
		size="small"
		:platform="deviceStore.platform"
		@onSubmit="onConfirm"
	>
		<div class="unbind-app__body">
			<i18n-t
				:keypath="messageKeypath"
				tag="div"
				scope="global"
				class="text-ink-2"
				:class="deviceStore.isMobile ? 'text-body3-m' : 'text-body3'"
			>
				<template #app>
					<span class="unbind-app__name">{{ appTitle }}</span>
				</template>
				<template #card>
					<span class="unbind-app__name">{{ cardName }}</span>
				</template>
			</i18n-t>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDeviceStore } from 'src/stores/settings/device';

/**
 * Presentation-only confirm dialog shared by Settings and Market. Callers pass
 * a localized `title` and `messageKeypath` from their own i18n namespace; the
 * message may reference `{app}` and `{card}` slots (rendered emphasized).
 */
defineProps<{
	title: string;
	messageKeypath: string;
	appTitle: string;
	cardName?: string;
	okText?: string;
}>();

const { t } = useI18n();
const deviceStore = useDeviceStore();
const customRef = ref();

const onConfirm = () => {
	customRef.value?.onDialogOK();
};
</script>

<style scoped lang="scss">
.unbind-app__body {
	padding: 0;
	word-wrap: break-word;
	white-space: normal;

	.unbind-app__name {
		color: $ink-1;
		font-weight: 500;
	}
}
</style>
