<template>
	<div class="terminus-unlock-box column justify-start">
		<q-item class="q-pa-none q-mt-md">
			<q-avatar :size="`${44}px`">
				<TerminusAvatar :info="tokenStore.user" :size="44" />
			</q-avatar>
			<div class="justify-between q-ml-md">
				<div class="text-subtitle1 text-ink-1">
					{{ tokenStore.user?.olaresId?.split('@')[0] }}
				</div>
				<div class="text-body3 text-ink-2">
					{{ tokenStore.user?.olaresId }}
				</div>
			</div>
		</q-item>
		<terminus-edit
			v-model="passwordRef"
			:show-password-img="true"
			class="terminus-unlock-box__edit"
			@update:model-value="onTextChange"
			:isError="!!message"
			:errorMessage="message"
			@keyup.enter="onSubmit"
		/>
		<confirm-button
			class="terminus-unlock-box__button"
			:btn-status="btnStatusRef"
			:btn-title="t('unlock.title')"
			@onConfirm="onSubmit()"
		/>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../../components/common/TerminusEdit.vue';
import ConfirmButton from '../../components/common/ConfirmButton.vue';
import { ref } from 'vue';
import { ConfirmButtonStatus } from '../../utils/constants';
import { useTokenStore } from 'src/stores/share/token';
const props = defineProps({
	message: {
		type: String,
		required: false,
		default: ''
	}
});

const tokenStore = useTokenStore();

const { t } = useI18n();
const passwordRef = ref('');
const btnStatusRef = ref<ConfirmButtonStatus>(ConfirmButtonStatus.disable);

function onTextChange() {
	emit('resetError');
	btnStatusRef.value =
		passwordRef.value.length > 0
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
}

async function onSubmit() {
	if (btnStatusRef.value === ConfirmButtonStatus.disable) return;
	emit('login', passwordRef.value);
}

const emit = defineEmits(['login', 'resetError']);
</script>

<style scoped lang="scss">
.terminus-unlock-box {
	width: 400px;
	border-radius: 12px;
	padding: 20px;
	background: $background-2;

	&__desc {
		margin-top: 12px;
	}

	&__edit {
		margin-top: 20px;
		width: 100%;
	}

	&__button {
		margin-top: 30px;
		width: calc(100%);
	}
}
</style>
