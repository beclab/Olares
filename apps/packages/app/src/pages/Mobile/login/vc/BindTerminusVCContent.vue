<template>
	<div class="bind-terminus-content-root">
		<div :class="!hasBtn ? 'full_area' : 'content-scroll-area'">
			<q-scroll-area
				class="terminus_bind_base_root"
				content-style="width: 100%;"
			>
				<div
					class="bind-terminus-content-root-content column justify-start items-center"
				>
					<q-img
						class="terminus_bind_base_root__logo"
						:src="getRequireImage('login/create_terminus_name.svg')"
					/>
					<div class="terminus_bind_base_root__desc text-body2">
						<slot name="desc"></slot>
					</div>
					<div class="terminus_bind_base_root__content">
						<slot name="content"></slot>
					</div>
				</div>
			</q-scroll-area>
		</div>
		<div class="btn-part" v-if="hasBtn">
			<confirm-button
				:btn-title="btnTitle"
				@onConfirm="onConfirm"
				@onError="onError"
				:btn-status="btnStatus"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getRequireImage } from '../../../../utils/imageUtils';
import './bindTerminusVCCommon.scss';
import { i18n } from '../../../../boot/i18n';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import ConfirmButton from '../../../../components/common/ConfirmButton.vue';
import { onMounted, ref } from 'vue';
import { useQuasar } from 'quasar';
import MonitorKeyboard from '../../../../utils/monitorKeyboard';
import { onUnmounted } from 'vue';

defineProps({
	hasBtn: {
		type: Boolean,
		default: false,
		required: false
	},
	btnTitle: {
		type: String,
		required: false,
		default: i18n.global.t('confirm')
	},
	btnStatus: {
		type: Number,
		require: false,
		default: ConfirmButtonStatus.normal
	},
	btnIcon: {
		type: String,
		required: false,
		default: ''
	}
});

const $q = useQuasar();

let monitorKeyboard: MonitorKeyboard | undefined = undefined;

const keyboardOpen = ref(false);

onMounted(() => {
	if ($q.platform.is.android) {
		monitorKeyboard = new MonitorKeyboard();
		monitorKeyboard.onStart();
		monitorKeyboard.onShow(() => (keyboardOpen.value = true));
		monitorKeyboard.onHidden(() => (keyboardOpen.value = false));
	}
});

onUnmounted(() => {
	if ($q.platform.is.android) {
		if (monitorKeyboard) {
			monitorKeyboard.onEnd();
		}
	}
});

const emit = defineEmits(['onConfirm', 'onError']);

const onConfirm = () => {
	emit('onConfirm');
};

const onError = () => {
	emit('onError');
};
</script>

<style scoped lang="scss">
.bind-terminus-content-root {
	width: 100%;
	height: 100%;

	.full_area {
		position: relative;
		height: 100%;
		width: 100%;
	}

	.content-scroll-area {
		position: relative;
		height: calc(100% - 88px);
		width: 100%;
	}

	.bind-terminus-content-root-content {
		width: 100%;
		height: 100%;
		padding-left: 20px;
		padding-right: 20px;
	}

	.btn-part {
		padding-left: 20px;
		padding-right: 20px;
	}
}
</style>
