<template>
	<div>
		<div class="mnemonics_wrap">
			<terminus-mnemonics-component
				ref="selectMnemonicsView"
				:readonly="true"
				:show-title="false"
				:is-backup="false"
				:is-copy="true"
				:is-paste="false"
				:mnemonics="mnemonic"
			/>
			<div class="mnemonics_login" v-if="encrypting">
				<q-icon
					name="sym_r_visibility_off"
					class="text-ink-on-brand"
					size="26px"
				/>
				<div class="q-mt-md q-ml-md q-mr-md content">
					{{ $t('back_up_your_mnemonic_phrase_immediately_to_safe') }}
				</div>
				<TerminusEnterBtn
					@sure-action="openCheckLogin"
					class="q-mt-lg"
					:title="$t('click_to_view')"
				/>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { onBeforeUnmount, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useUserStore } from '../../../stores/user';
import DialogLogin from './DialogLogin.vue';
import TerminusMnemonicsComponent from 'src/components/common/TerminusMnemonicsComponent.vue';
import TerminusEnterBtn from 'src/components/common/TerminusEnterBtn.vue';
import { busEmit } from '../../../utils/bus';
import { show, encrypting, hide } from './useMnemonics';

const $q = useQuasar();
const userStore = useUserStore();

const selectMnemonicsView = ref();

const mnemonicItem = userStore.current_mnemonic;

const mnemonic = ref(mnemonicItem?.mnemonic || '');
const openCheckLogin = async () => {
	if (!(await userStore.unlockFirst())) {
		return;
	}
	//TODO: 助记词写入逻辑前置
	if (!mnemonic.value) {
		mnemonic.value = userStore.current_mnemonic?.mnemonic || '';
		selectMnemonicsView.value.reloadMnemonics(mnemonic.value.split(' '));
	}
	if (!userStore.passwordReseted) {
		busEmit('configPassword');
		return;
	}
	if (process.env.APPLICATION_SUB_IS_BEX) {
		show();
		return;
	}
	$q.dialog({
		component: DialogLogin
	}).onOk(show);
};
onBeforeUnmount(() => {
	hide();
});
</script>

<style scoped lang="scss">
.mnemonics_wrap {
	position: relative;

	.mnemonics_login {
		width: 100%;
		height: 100%;
		position: absolute;
		top: 0;
		left: 0;
		fill: rgba(31, 24, 20, 0.6);
		backdrop-filter: blur(8px);
		background: rgba(0, 0, 0, 0.5);
		border-radius: 8px;
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 0 20px;
		color: $white;

		.content {
			line-height: 24px;
			text-align: center;
		}

		.click {
			padding: 8px;
			border-radius: 8px;
			background: $yellow;
			color: $grey-10;
			cursor: pointer;

			&:hover {
				background: $yellow-13;
			}
		}
	}
}
</style>
