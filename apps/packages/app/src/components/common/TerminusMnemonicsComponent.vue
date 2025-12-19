<template>
	<div class="mnemonics-root">
		<div
			v-if="showTitle && lagerTitle"
			class="text-color-title text-h6"
			style="margin-bottom: 12px"
		>
			{{ t('mnemonic_phrase') }}
		</div>
		<div v-else-if="showTitle" class="login-label q-mb-xs">
			{{ t('mnemonic_phrase') }}
		</div>
		<div class="mnemonics-bg row justify-between" v-if="!forceReloadInput">
			<terminus-mnemonic-item
				v-for="(item, index) in mnemonicsRef"
				:key="index"
				:input-text="item.text"
				:is-read-only="readonly"
				:is-backup="isBackup"
				:backup-readonly="backupReadonly"
				:is-error="item.isError"
				:class="calcMnemonicPositiionClasses(index)"
				:style="`width: calc((100%) / ${rowCapacity})`"
				:index="index"
				@on-text-change="textOnChanged"
				@on-finished-edit="onFishedEdit"
				@on-backup-delete-item="onBackupDeleteItem"
				@on-update-error="errorOnChanged"
			/>
		</div>

		<div
			class="row items-center justify-center"
			v-if="(isCopy || isPaste) && (!isBex || (isBex && isCopy))"
		>
			<div
				class="paste-parent q-mt-lg row justify-center items-center"
				:class="isPaste ? 'text-blue-4' : 'text-ink-1'"
				@click="pasteFunc"
			>
				<q-icon size="16px" name="sym_r_content_copy" v-if="isCopy" />
				<q-icon size="16px" name="sym_r_content_paste" v-if="isPaste" />
				<div class="paste text-body3" v-if="isPaste">
					{{ t('paste') }}
				</div>
				<div class="paste text-body3" v-if="isCopy">
					{{ t('copy') }}
				</div>
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { getPlatform } from '@didvault/sdk/src/core';
import { nextTick, ref, watch } from 'vue';
import TerminusMnemonicItem from '../../components/common/TerminusMnemonicItem.vue';
import { walletService } from '../../wallet';
import { useI18n } from 'vue-i18n';

import { notifySuccess } from '../../utils/notifyRedefinedUtil';

const { t } = useI18n();

interface MnemonicInterface {
	text: string;
	isError: boolean;
}

const props = defineProps({
	lagerTitle: {
		type: Boolean,
		default: true
	},
	mnemonics: {
		type: String,
		required: false,
		default: ''
	},
	readonly: {
		type: Boolean,
		required: false,
		default: false
	},
	maxLength: {
		type: Number,
		required: false,
		default: 12
	},
	isBackup: {
		type: Boolean,
		required: false,
		default: false
	},
	showTitle: {
		type: Boolean,
		required: false,
		default: true
	},

	isCopy: {
		type: Boolean,
		required: false,
		default: false
	},

	isPaste: {
		type: Boolean,
		required: false,
		default: true
	},

	backupReadonly: {
		type: Boolean,
		required: false,
		default: true
	},
	rowCapacity: {
		type: Number,
		required: false,
		default: 3
	}
});

const isBex = ref(process.env.IS_BEX);

const mnemonicsLength = ref(
	props.mnemonics.length > 0
		? props.mnemonics.split(' ').length > props.maxLength
			? props.maxLength
			: props.mnemonics.split(' ').length
		: props.maxLength
);

const mnemonicsRef = ref<MnemonicInterface[]>([]);

const mnemonicsRootRef = ref(props.mnemonics);

const updateMnemonicsRootRef = (value) => {
	mnemonicsRootRef.value = value;
};

const mnemonicsInit = () => {
	const mnemonics: MnemonicInterface[] = [];
	const defaultMnemonics: string[] =
		mnemonicsRootRef.value.length > 0 ? mnemonicsRootRef.value.split(' ') : [];
	for (let index = 0; index < mnemonicsLength.value; index++) {
		const text =
			defaultMnemonics.length >= index + 1 ? defaultMnemonics[index] : '';
		mnemonics.push({
			text: text,
			isError:
				props.readonly || text.length === 0 || props.isBackup
					? false
					: !walletService.walletCore.Mnemonic.isValidWord(text)
		});
	}
	mnemonicsRef.value = mnemonics;
};

const calcMnemonicPositiionClasses = (index: number) => {
	let mnemonicClassArray: string[] = [];

	if (parseInt((index / props.rowCapacity).toString()) > 0) {
		mnemonicClassArray.push('mnemonic-top');
	}

	if (index % props.rowCapacity > 0) {
		mnemonicClassArray.push('mnemonic-left');
	}

	return mnemonicClassArray.join(' ');
};

const forceReloadInput = ref(false);

const textOnChanged = (
	index: number,
	text: string,
	next = false,
	nextName = ''
) => {
	const texts = text.split(' ');
	if (texts.length <= 1) {
		if (next) {
			const aaa = document.getElementsByName(nextName);
			if (aaa.length > 0) {
				aaa[0].focus();
			}
		}
		setIndexText(index, text);
		return;
	}

	forceReloadInput.value = true;
	for (
		let index = 0;
		index < Math.min(mnemonicsLength.value, texts.length);
		index++
	) {
		setIndexText(index, texts[index]);
		onFishedEdit(index, texts[index]);
	}
	nextTick(() => {
		forceReloadInput.value = false;
	});
};

const setIndexText = (index: number, text: string) => {
	if (index + 1 <= mnemonicsLength.value) {
		mnemonicsRef.value[index] = {
			text: text,
			isError: false
		};
	}
};

const errorOnChanged = (index: number, status: boolean) => {
	mnemonicsRef.value[index].isError = status;
};

const onFishedEdit = (index: number, text: string) => {
	const isError =
		text.length === 0 || props.isBackup
			? false
			: !walletService.walletCore.Mnemonic.isValidWord(text.trim());
	mnemonicsRef.value[index] = {
		text: text,
		isError: isError
	};

	emit('onMnemonicChange', mnemonicsRef.value.map((e) => e.text).join(' '));
};

const pasteFunc = async () => {
	if (props.isPaste) {
		try {
			let text = await getPlatform().getClipboard();
			text = text.replace(/(^\s*)|(\s*$)/g, '');
			if (text.length > 0 && text.split(' ').length === mnemonicsLength.value) {
				textOnChanged(0, text);
			}
		} catch (error) {
			console.error(error.message);
		}
	} else {
		try {
			await getPlatform().setClipboard(props.mnemonics);
			notifySuccess(t('copy_success'));
		} catch (error) {
			console.error(error.message);
		}
	}
};

const emit = defineEmits(['onMnemonicChange', 'backupRemoveItem']);

const clearInput = () => {
	forceReloadInput.value = true;
	mnemonicsInit();
	nextTick(() => {
		forceReloadInput.value = false;
		emit('onMnemonicChange', mnemonicsRef.value.map((e) => e.text).join(' '));
	});
};

const reloadMnemonics = (menmonics: string[]) => {
	for (let index = 0; index < menmonics.length; index++) {
		const mnemonic = menmonics[index];
		mnemonicsRef.value[index] = {
			text: mnemonic,
			isError: false
		};
	}
	forceReloadInput.value = true;
	nextTick(() => {
		forceReloadInput.value = false;
	});
};

const onBackupDeleteItem = (index: number) => {
	mnemonicsRef.value[index] = {
		text: '',
		isError: false
	};
	forceReloadInput.value = true;

	emit('backupRemoveItem', index);

	nextTick(() => {
		forceReloadInput.value = false;
	});
};

watch(
	() => props.mnemonics,
	(newValue) => {
		updateMnemonicsRootRef(newValue);
		mnemonicsInit();
	},
	{
		immediate: true
	}
);

defineExpose({ clearInput, reloadMnemonics });
</script>

<style scoped lang="scss">
.mnemonics-root {
	width: 100%;

	.mnemonics-bg {
		width: 100%;

		.mnemonic-left {
			padding-left: 13px;
		}

		.mnemonic-top {
			margin-top: 12px;
		}
	}

	.paste-parent {
		padding: 4px 8px;
		width: auto;
		border-radius: 4px;
		border: 1px solid $separator;

		.paste {
			margin-left: 4px;
		}
	}
}
</style>
