<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:skip="false"
		:ok="confirmText"
		:platform="deviceStore.platform"
		:cancel="useCancel ? cancelText : false"
		:size="affectedDomains.length > 0 ? 'medium' : 'small'"
		@onSubmit="sureAction"
	>
		<div class="row items-center justify-between" v-if="message.length > 0">
			<div
				class="text-ink-2"
				:class="{
					'message-content': hasBorder,
					'text-body3': !deviceStore.isMobile,
					'text-body3-m': deviceStore.isMobile
				}"
				v-html="message"
				:style="deviceStore.isMobile ? 'text-align: center' : ''"
			/>
		</div>
		<div v-if="affectedDomains.length > 0">
			<div class="q-mt-xl text-body1 text-ink-3">
				{{ t('Affected domain') }}
			</div>
			<div
				class="affection-domain-bg text-body1 text-ink-2 bg-background-6 row items-center q-pl-md q-mt-sm"
				v-for="(domain, index) in affectedDomains"
				:key="index"
			>
				<q-icon name="sym_r_language" size="18px" />
				<div class="q-ml-sm">
					{{ domain.custom_domain }}
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { i18n } from 'src/boot/i18n';
import { useDeviceStore } from 'src/stores/settings/device';
import { ApplicationCustonDomain } from 'src/constant/global';
import { useI18n } from 'vue-i18n';

defineProps({
	title: {
		type: String,
		required: false,
		default: ''
	},
	message: {
		type: String,
		required: false,
		default: ''
	},
	useCancel: {
		type: Boolean,
		required: false,
		default: true
	},
	confirmText: {
		type: String,
		required: false,
		default: i18n.global.t('ok')
	},
	cancelText: {
		type: String,
		required: false,
		default: i18n.global.t('cancel')
	},
	hasBorder: {
		type: Boolean,
		required: false,
		default: false
	},
	isReminder: {
		type: Boolean,
		required: false,
		default: false
	},
	affectedDomains: {
		type: Array as () => Array<ApplicationCustonDomain>,
		default: [] as ApplicationCustonDomain[]
	}
});

const CustomRef = ref();

const { t } = useI18n();

const sureAction = () => {
	CustomRef.value.onDialogOK();
};

const deviceStore = useDeviceStore();
</script>

<style scoped lang="scss">
.content-root {
	padding: 16px 16px 0;
}

.userName {
	line-height: 32px;
	height: 32px;
}

.message-content {
	width: 100%;
	border: 1px solid $separator;
	padding: 8px 12px;
	border-radius: 8px;
}

.affection-domain-bg {
	height: 40px;
	border-radius: 8px;
}
</style>
