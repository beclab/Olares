<template>
	<div class="backup-mnemonic-root">
		<TerminusMobileTitleConfirmView
			:title="
				isMnemonicBackup ? t('backup_mnemonic_phrase') : t('mnemonic_phrase')
			"
			:btn-title="stepMap[step]"
			:btn-status="btnStatusRef"
			:btn-hide="!isMnemonicBackup"
			@on-confirm="confirmAction"
		>
			<template v-slot:content>
				<div class="backup-content">
					<div v-if="isMnemonicBackup">
						<div class="step-view row items-center justify-between">
							<div
								v-for="index in 3"
								:key="index"
								class="step-view__item"
								:class="
									index <= step ? 'step-view__active' : 'step-view__normal'
								"
								:style="`width:calc((100% - 30px - 20px) / ${3});margin-left:${
									index > 1 ? '10' : '0'
								}px;`"
							></div>
						</div>
						<div class="step-content">
							<div v-if="step == BackupStep.start">
								<div class="text-h6 text-color-title">
									{{ t('please_clarify_the_information') }}
								</div>

								<!--
							<div
								class="module-introduce text-body2 q-mt-lg"
								v-html="reminderInfoRef"
							></div> -->

								<div
									class="module-reminder q-mt-lg q-pl-md q-pr-md q-pt-md q-pb-md"
								>
									<div class="row module-reminder__item">
										<q-icon name="sym_r_error" color="negative" size="20px" />
										<div class="module-reminder__text text-body2 q-ml-sm">
											{{ t('without_the_mnemonic_there_is_no_olares_id') }}
										</div>
									</div>
									<div class="row module-reminder__item q-mt-md">
										<q-icon name="sym_r_error" color="negative" size="20px" />
										<div class="module-reminder__text text-body2 q-ml-sm">
											{{ t('without_olares_id_there_is_no_data_for_you') }}
										</div>
									</div>
								</div>
							</div>
							<div v-else-if="step == BackupStep.display">
								<div class="text-h6 text-color-title">
									{{ t('copy_mnemonics') }}
								</div>

								<div class="q-mt-md text-body3 module-sub-title">
									{{ t('mnemonic_phrase') }}
								</div>
								<div class="q-mt-sm mnemonics_bg">
									<terminus-mnemonics-component
										:readonly="true"
										:show-title="false"
										:is-backup="true"
										:mnemonics="mnemonic"
										:is-copy="true"
										:is-paste="false"
									/>

									<div class="mnemonics_login" v-if="encrypting">
										<q-icon name="visibility_off" />
										<div class="q-my-md text-body3 content">
											{{ t('backup_mnemonics_reminder_info') }}
										</div>
										<div class="click" @click="openCheckLogin">
											{{ t('click_to_view') }}
										</div>
									</div>
								</div>
							</div>
							<div v-else-if="step == BackupStep.select">
								<div class="text-h6 text-color-title">
									{{
										t(
											'please_select_each_word_in_the_order_presented_previously'
										)
									}}
								</div>
								<div class="q-mt-md text-body3 module-sub-title">
									{{ t('mnemonic_phrase') }}
								</div>
								<div class="q-mt-sm" style="width: calc(100% - 8px)">
									<terminus-mnemonics-component
										ref="selectMnemonicsView"
										:readonly="true"
										:show-copy="false"
										:show-title="false"
										:is-backup="true"
										:backup-readonly="false"
										:is-paste="false"
										:mnemonics="selectMnemonics.join(' ')"
										@backup-remove-item="backupRemoveItem"
									/>
								</div>
								<div class="row q-mr-md q-mt-xl">
									<div
										v-for="(mnemonic, index) in randomMnemonics"
										@click="itemOnClick(index)"
										:key="index"
										class="q-px-md q-py-xs q-mr-md q-mb-md mnemonic-item text-body2"
									>
										{{ mnemonic }}
									</div>
								</div>
							</div>
						</div>
					</div>
					<div v-else>
						<div class="q-mt-md text-body3 module-sub-title">
							{{ t('mnemonic_phrase') }}
						</div>
						<div class="q-mt-sm mnemonics_bg">
							<terminus-mnemonics-component
								:readonly="true"
								:show-title="false"
								:is-backup="false"
								:mnemonics="mnemonic"
								:is-copy="true"
								:is-paste="false"
							/>

							<div class="mnemonics_login" v-if="encrypting">
								<q-icon name="visibility_off" />
								<div class="q-my-md text-body3 content">
									{{ t('backup_mnemonics_reminder_info') }}
								</div>
								<div class="click" @click="openCheckLogin">
									{{ t('click_to_view') }}
								</div>
							</div>
						</div>
					</div>
				</div>
			</template>
		</TerminusMobileTitleConfirmView>
	</div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import TerminusMobileTitleConfirmView from '../../../components/common/TerminusMobileTitleConfirmView.vue';
import TerminusMnemonicsComponent from '../../../components/common/TerminusMnemonicsComponent.vue';
import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { ConfirmButtonStatus } from '../../../utils/constants';
import { ref } from 'vue';
import { useUserStore } from '../../../stores/user';
import { watch } from 'vue';
import { useQuasar } from 'quasar';

import DialogLogin from 'src/pages/Electron/SettingsPage/DialogLogin.vue';
import { useRouter, useRoute } from 'vue-router';

enum BackupStep {
	start = 1,
	display,
	select
}

const { t } = useI18n();

const stepMap: Record<BackupStep, string> = {
	[BackupStep.start]: t('start'),
	[BackupStep.display]: t('buttons.next'),
	[BackupStep.select]: t('completed')
};

const step = ref(BackupStep.start);

const btnStatusRef = ref(ConfirmButtonStatus.normal);

const userStore = useUserStore();

const currentUser = userStore.users?.items.get(userStore.current_id || '');

const $router = useRouter();

const encrypting = ref(true);

const $q = useQuasar();

const mnemonic = ref(userStore.current_mnemonic?.mnemonic || '');

const mnemonicsRef = ref([] as string[]);
const randomMnemonics = ref([] as string[]);
const selectMnemonics = ref([] as string[]);

const mnemonicsInit = () => {
	mnemonicsRef.value = mnemonic.value.split(' ');
	let leftMnemonics = mnemonic.value.split(' ');
	selectMnemonics.value = mnemonicsRef.value.map(() => '');
	while (randomMnemonics.value.length < mnemonicsRef.value.length) {
		if (leftMnemonics.length == 1) {
			randomMnemonics.value.push(...leftMnemonics);
			break;
		}
		const index = randomNum(1, leftMnemonics.length);
		const t = leftMnemonics.splice(index, 1);
		randomMnemonics.value.push(...t);
	}
};

const confirmAction = async () => {
	if (step.value <= BackupStep.display) {
		step.value = step.value + 1;
		btnStatusRef.value = ConfirmButtonStatus.disable;
		return;
	}

	if (selectMnemonics.value.join(' ') !== mnemonic.value) {
		notifyFailed(t('mnemonics_no_match'));
		return;
	}

	await userStore.backupCurrentUser();

	$router.back();
};

watch(
	randomMnemonics,
	() => {
		if (step.value == BackupStep.select) {
			if (randomMnemonics.value.length == 0) {
				btnStatusRef.value = ConfirmButtonStatus.normal;
			} else {
				btnStatusRef.value = ConfirmButtonStatus.disable;
			}
		}
	},
	{ immediate: true, deep: true }
);

const openCheckLogin = () => {
	$q.dialog({
		component: DialogLogin
	}).onOk(() => {
		encrypting.value = false;
		btnStatusRef.value = ConfirmButtonStatus.normal;
	});
};

const randomNum = (minNum: number, maxNum: number) => {
	return Math.floor(Math.random() * maxNum) + minNum;
};

mnemonicsInit();

const selectMnemonicsView = ref();

const itemOnClick = (index: number) => {
	const t = randomMnemonics.value.splice(index, 1);
	for (let position = 0; position < selectMnemonics.value.length; position++) {
		if (selectMnemonics.value[position] == '') {
			selectMnemonics.value[position] = t[0];
			break;
		}
	}
	selectMnemonicsView.value.reloadMnemonics(selectMnemonics.value);
};

const backupRemoveItem = (index: number) => {
	const t = selectMnemonics.value[index];
	randomMnemonics.value.push(t);
	selectMnemonics.value[index] = '';
};

const route = useRoute();

const backup = route.query.backup ? (route.query.backup as string) : '';
console.log('route.query ====>', backup);

const isMnemonicBackup = ref(
	!process.env.IS_BEX &&
		(backup === '1' ||
			(!userStore.currentUserBackup && currentUser?.setup_finished))
);
</script>

<style scoped lang="scss">
.backup-mnemonic-root {
	width: 100%;
	height: 100%;

	.backup-content {
		padding-left: 20px;
		padding-right: 20px;
		width: 100%;
	}

	.step-view {
		height: 30px;
		width: 100%;

		&__item {
			height: 4px;
		}

		&__active {
			background-color: $yellow;
		}

		&__normal {
			background-color: $grey-12;
		}
	}

	.module-introduce {
		color: $grey-8;
	}

	.module-reminder {
		border: 1px solid $red;
		border-radius: 8px;
		width: 100%;
		background-color: $red-1;

		&__item {
			width: 100%;
		}

		&__text {
			color: $red;
			max-width: calc(100% - 40px);
		}
	}

	.step-content {
		width: 100%;

		.mnemonic-item {
			border: 1px solid $yellow;
			border-radius: 4px;
		}
	}

	.module-sub-title {
		text-align: left;
		color: $prompt-message;
		text-transform: capitalize;
	}

	.mnemonics_bg {
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
				text-align: center;
				color: $white;
			}

			.click {
				padding: 8px;
				border-radius: 8px;
				background: $yellow;
				color: $grey-10;
				cursor: pointer;

				&:hover {
					background: $yellow;
				}
			}
		}
	}
}
</style>
