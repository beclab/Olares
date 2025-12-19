<template>
	<div class="row items-center justify-end header-content">
		<div v-if="$slots.back" @click="backAction">
			<slot name="back" />
		</div>
		<q-space />
		<div
			v-if="scan"
			class="scan-icon row items-center justify-center"
			@click="scanQrCode"
		>
			<q-img
				:src="
					$q.dark.isActive
						? getRequireImage('common/dark_scan_qr.svg')
						: getRequireImage('common/scan_qr.svg')
				"
				width="20px"
			/>
		</div>

		<div
			v-if="$slots.avatar"
			@click="enterAccounts"
			class="avatar-margin-right row items-center justify-center"
		>
			<slot name="avatar" />
		</div>
		<div v-else class="avatar avatar-margin-right q-ml-xs">
			<TerminusAvatar
				:info="userStore.terminusInfo()"
				:size="40"
				@click="enterAccounts"
			/>
		</div>
	</div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router';
import { useUserStore } from '../../stores/user';
import { getRequireImage } from 'src/utils/imageUtils';
import { useQuasar } from 'quasar';
import SwitchAccount from '../SwitchAccount.vue';

const props = defineProps({
	scan: {
		type: Boolean,
		required: false,
		default: true
	},
	back: {
		type: Boolean,
		required: false,
		default: false
	},
	redefinedAvatar: {
		type: Boolean,
		required: false,
		default: false
	}
});

const userStore = useUserStore();

const router = useRouter();

const $q = useQuasar();

const enterAccounts = () => {
	if (props.redefinedAvatar) {
		emits('redefinedAvatarAction');
		return;
	}
	if (process.env.PLATFORM == 'DESKTOP' || process.env.APPLICATION_SUB_IS_BEX) {
		handleSwitchAccount();
		return;
	}
	router.push('/accounts');
};

const handleSwitchAccount = () => {
	$q.dialog({
		component: SwitchAccount
	});
};

const backAction = () => {
	router.back();
};

const scanQrCode = () => {
	if (!props.scan) {
		return;
	}
	router.push({
		path: '/scanQrCode'
	});
};
const emits = defineEmits(['redefinedAvatarAction']);
</script>

<style scoped lang="scss">
.header-content {
	width: 100%;
	.scan-icon {
		width: 40px;
		height: 40px;
	}
}
.avatar {
	height: 40px;
	width: 40px;
	overflow: hidden;
	border-radius: 20px;
}
.avatar-margin-right {
	margin-right: 20px;
	height: 40px;
	width: 40px;
}
</style>
