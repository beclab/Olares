<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="title"
		:ok="false"
		:cancel="false"
		:size="'small'"
		:platform="'mobile'"
	>
		<UserStatusCommonContent
			@on-dialog-o-k="onDialogOK"
			:btn-title="btnTitle"
			:btnRedefined="isReactive || addSkip || setDomain"
			:message="message"
			:messageClasses="messageClasses ? messageClasses : undefined"
			:messageCenter="messageCenter"
		>
		</UserStatusCommonContent>
		<template v-slot:footerMore v-if="isReactive">
			<div class="column items-center justify-between" style="width: 100%">
				<div
					class="reactivate-mode text-subtitle1 row items-center justify-center q-mb-md"
					@click="reactivateAction"
				>
					{{ t('reactivate') }}
				</div>

				<div
					class="offline-mode text-subtitle1 text-ink-1 row items-center justify-center q-mb-md"
					@click="offLineAction"
				>
					{{ t('user_current_status.offline_mode.enable') }}
				</div>

				<div
					class="offline-mode text-subtitle1 text-ink-1 row items-center justify-center"
					@click="onDialogOK"
				>
					{{ btnTitle }}
				</div>
			</div>
		</template>
		<template v-slot:footerMore v-else-if="addSkip">
			<div class="column items-center justify-between" style="width: 100%">
				<div
					class="reactivate-mode text-subtitle1 row items-center justify-center q-mb-md"
					@click="onDialogOK"
				>
					{{ btnTitle }}
				</div>

				<div
					class="skip text-subtitle1 row items-center justify-center"
					@click="onDialogCancel"
				>
					{{ skipTitle }}
				</div>
			</div>
		</template>
		<template v-slot:footerMore v-else-if="setDomain">
			<div class="column items-center justify-between" style="width: 100%">
				<div
					class="resetDomain row items-center justify-center q-mb-md"
					:class="resetDomainClasses"
					@click="onDialogOK('reset')"
				>
					{{ resetDomainTitle }}
				</div>
				<div
					class="reactivate-mode text-subtitle1 row items-center justify-center"
					@click="onDialogOK('confirm')"
				>
					{{ btnTitle }}
				</div>
			</div>
		</template>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import UserStatusCommonContent from './UserStatusCommonContent.vue';
import { useUserStore } from '../../stores/user';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../boot/i18n';
import { ref } from 'vue';
const userStore = useUserStore();

const $router = useRouter();

const { t } = useI18n();

defineProps({
	title: String,
	message: String,
	btnTitle: String,
	doubleBtn: {
		default: false,
		type: Boolean,
		required: false
	},
	isReactive: {
		default: false,
		type: Boolean,
		required: false
	},
	addSkip: {
		default: false,
		type: Boolean,
		required: false
	},
	skipTitle: {
		type: String,
		default: i18n.global.t('skip'),
		required: false
	},
	messageClasses: {
		type: String,
		default: '',
		required: false
	},
	messageCenter: {
		default: false,
		type: Boolean,
		required: false
	},
	setDomain: {
		default: false,
		type: Boolean,
		required: false
	},
	resetDomainTitle: {
		type: String,
		default: '',
		required: false
	},
	resetDomainClasses: {
		type: String,
		default: 'text-subtitle1'
	}
});

const CustomRef = ref();

const offLineAction = () => {
	userStore.updateOfflineMode(true);
	CustomRef.value.onDialogOK();
};

const reactivateAction = () => {
	//
	$router.push({
		path: `/Activate/${1}`
	});
	CustomRef.value.onDialogOK();
};

const onDialogOK = (item: any) => {
	CustomRef.value.onDialogOK(item);
};

const onDialogCancel = () => {
	CustomRef.value.onDialogCancel();
};
</script>

<style lang="scss" scoped>
.q-dialog__backdrop {
	background: rgba(0, 0, 0, 0.7);
}

.offline-mode {
	border: 1px solid $separator;
	width: 100%;
	height: 48px;
	text-align: center;
	color: $ink-1;
	border-radius: 8px;
}

.reactivate-mode {
	border: 1px solid $yellow;
	width: 100%;
	height: 48px;

	text-align: center;
	color: $grey-10;
	border-radius: 8px;
	background: $yellow;
}

.skip {
	height: 48px;
	text-align: center;
	color: $light-blue;
	width: 100%;
}

.resetDomain {
	height: 48px;
	text-align: center;
	// color: $light-blue;
	width: 100%;
}
</style>
