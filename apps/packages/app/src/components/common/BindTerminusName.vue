<template>
	<div
		class="bind-name"
		:style="
			showTypeRef === BIND_STATUS.BIND_OK || showTypeRef === BIND_STATUS.VC_NONE
				? 'height : auto;'
				: ''
		"
	>
		<v-c-empty-view
			style="width: 100%"
			v-if="
				showTypeRef === BIND_STATUS.VC_NONE ||
				showTypeRef === BIND_STATUS.TERMINUS_NAME_NONE
			"
		/>
		<div
			class="column justify-center items-center"
			style="width: 100%; height: 100%"
			v-if="showTypeRef === BIND_STATUS.OS_NONE"
		>
			<terminus-avatar
				:did="userStore.terminusInfo()"
				:size="64"
				class="avatar-circle"
			/>
			<div class="text-color-title text-subtitle2 bind-name__text_name">
				{{ terminusNameRef }}
			</div>
			<div class="bind-name__text_desc text-body3 text-ink-2">
				{{ t('not_connected_olares_log_in_now') }}
			</div>
			<TerminusEnterBtn
				:title="t('scan_the_qr_code')"
				icon="sym_r_qr_code_scanner"
				@sureAction="
					needBindStatusOkToNext(router, $q.platform.is.nativeMobile || false)
				"
				class="q-mt-md"
			/>
		</div>

		<div
			class="column justify-center items-center"
			style="width: 100%; height: 100%"
			v-if="showTypeRef === BIND_STATUS.OS_WAIT_DNS_OK"
		>
			<terminus-avatar
				:info="userStore.terminusInfo()"
				:size="64"
				class="avatar-circle"
			/>
			<div class="text-subtitle2 bind-name__text_name">
				{{ terminusNameRef }}
			</div>
			<div class="bind-name__text_desc text-body3">
				{{ t('wait_for_terminus_to_activate') }}
			</div>
		</div>

		<div
			style="width: 100%; height: 100%"
			class="column justify-center"
			v-if="$slots.success && showTypeRef === BIND_STATUS.BIND_OK"
		>
			<slot name="success" />
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import VCEmptyView from '../../pages/Mobile/items/VCEmptyView.vue';
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import { useQuasar } from 'quasar';
import {
	BIND_STATUS,
	current_user_bind_status
} from '../../utils/terminusBindUtils';
import { useI18n } from 'vue-i18n';
import TerminusEnterBtn from './TerminusEnterBtn.vue';
import { needBindStatusOkToNext } from '../../utils/terminusBindUtils';

defineProps({
	borderShow: {
		type: Boolean,
		default: true,
		required: false
	}
});

const { t } = useI18n();
const $q = useQuasar();
const router = useRouter();
const userStore = useUserStore();
const showTypeRef = ref<BIND_STATUS>();
const terminusNameRef = ref();

showTypeRef.value = current_user_bind_status();
terminusNameRef.value = userStore.current_user?.name;
</script>

<style scoped lang="scss">
.bind-name {
	height: 200px;
	width: 100%;

	&__text_name {
		text-align: center;
		margin-top: 8px;
		margin-left: 40px;
		margin-right: 40px;
	}

	&__text_desc {
		text-align: center;
		margin-top: 12px;
		margin-left: 40px;
		margin-right: 40px;
	}

	&__btn {
		margin-top: 11px;
		width: 126px;
		height: 32px;
	}
}

.border_show {
	backdrop-filter: blur(6px);
	border-radius: 8px;
}
</style>
