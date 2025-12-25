<template>
	<div class="mnemonics-root">
		<div class="mnemonics-bg row justify-between">
			<template v-for="(item, index) in mnemonicsRef" :key="index">
				<div class="terminus-mnemonic col-4 col-sm-3">
					<div class="mnemonics-item q-px-sm q-ma-sm">
						<div class="text-gery-8">{{ index + 1 }}.</div>
						<q-input
							ref="inputRef"
							autocomplete="off"
							:model-value="item"
							type="text"
							dense
							autogrow
							disable
							borderless
							:input-style="{
								textAlign: 'center'
							}"
							input-class="mnemonic-input text-ink-1"
							text-align="center"
						/>
					</div>
				</div>
			</template>
		</div>

		<div class="row items-center justify-center">
			<div class="copy text-body3" @click="copyFunc">
				<q-icon name="sym_r_content_copy" size="20px" />
				{{ t('Copy') }}
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
import { getPlatform } from '@didvault/sdk/src/core';
import {
	notifySuccess,
	notifyFailed
} from '../../../utils/notifyRedefinedUtil';

const props = defineProps({
	mnemonic: {
		type: String,
		required: false,
		default: ''
	}
});

const mnemonicsRef = ref();

const mnemonicsInit = () => {
	mnemonicsRef.value = props.mnemonic.split(' ');
};

onMounted(() => {
	mnemonicsInit();
});

const copyFunc = async () => {
	const platform = getPlatform();
	platform
		.setClipboard(props.mnemonic)
		.then(() => {
			notifySuccess(t('copy_success'));
		})
		.catch(() => {
			notifyFailed(t('copy_fail'));
		});
};
</script>

<style scoped lang="scss">
.mnemonics-root {
	width: 100%;

	.mnemonics-bg {
		width: 100%;
		margin-top: 8px;

		.terminus-mnemonic {
			.mnemonics-item {
				border-radius: 8px;
				border: 1px solid $separator;
				display: flex;
				align-items: center;
				justify-content: space-between;
			}
		}
	}

	.copy {
		text-align: center;
		margin-top: 10px;
		border-radius: 4px;
		border: 1px solid $separator;
		padding: 2px 6px;
		cursor: pointer;
	}
}
</style>
