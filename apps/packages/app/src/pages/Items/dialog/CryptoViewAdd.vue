<template>
	<q-dialog
		class="d-creatVault text-center"
		ref="dialogRef"
		@hide="onDialogHide"
	>
		<q-card class="q-dialog-plugin">
			<div
				class="column justify-center content-center items-center"
				style="height: 74px"
			>
				<div class="text-color-title text-h6 title">
					{{ t('crypto_wallet') }}
				</div>
			</div>

			<div class="stepHeader bg-grey-11">
				<div class="bg-blue" :style="{ width: stepProgress[step] }"></div>
			</div>

			<q-stepper v-model="step" color="primary" animated>
				<q-step :name="1" title="1" icon="settings" :done="step > 1">
					<div>
						<ButtonsSlider
							:options="options"
							ref="accountType"
							@on-change="onChildChange"
						></ButtonsSlider>
						<q-tab-panels v-model="tab" animated class="scroll-bg">
							<q-tab-panel :name="options[0].value" class="content-bg">
								<div class="input-area-bg">
									<q-input
										v-model="walletName"
										type="textarea"
										autogrow
										class="section-wallet-name"
										bg-color="transparent"
										borderless
										:label="t('wallet_name')"
									/>
								</div>
								<div class="input-area-bg">
									<q-input
										v-model="mnemonics"
										type="textarea"
										autogrow
										class="input-area"
										bg-color="transparent"
										borderless
										:label="t('mnemonic_phrase')"
										v-on:update:model-value="checkMnemonic()"
									/>
								</div>
								<div class="text-color-sub-detail text-body3 section-reminder">
									{{ t('vault_t.typically_64_alphanumeric_characters') }}
								</div>
							</q-tab-panel>

							<q-tab-panel name="privatekey" class="content-bg">
								<div class="input-area-bg">
									<q-input
										v-model="walletName"
										type="textarea"
										autogrow
										class="section-wallet-name"
										bg-color="transparent"
										borderless
										:label="t('wallet_name')"
									/>
								</div>
								<div class="input-area-bg">
									<q-input
										v-model="privatekey"
										type="textarea"
										:label="t('private_key')"
										autogrow
										class="input-area"
										bg-color="transparent"
										borderless
									/>
								</div>
								<div class="text-color-sub-detail text-body3 section-reminder">
									{{ t('vault_t.typically_64_alphanumeric_characters') }}
								</div>
							</q-tab-panel>
						</q-tab-panels>
					</div>
					<BottomConfirm
						:btn-title="t('import')"
						@confirmAction="onStep1()"
					></BottomConfirm>
				</q-step>

				<q-step :name="2" title="2" icon="settings" :done="step > 2">
					<div class="step2 row q-col-gutter-md">
						<div
							class="col-2 row items-center justify-center"
							v-for="(item, index) in walletOption"
							:key="index"
						>
							<BtCheckBox
								:name="item.name"
								:check="item.check"
								@toggle="toggleWallet(item.name)"
							/>
						</div>
					</div>

					<div class="row items-center justify-between q-my-md">
						<q-btn
							class="bg-grey-11 text-ink-2 text-body1 previous"
							:label="t('buttons.previous')"
							@click="onPrevious"
						/>
						<q-btn
							class="bg-blue text-body1 next"
							:label="t('buttons.next')"
							@click="onStep2"
						/>
					</div>
				</q-step>

				<q-step :name="3" title="3" icon="settings" :done="step > 3">
					<BtRadio
						:label="t('vault_t.show_in_sign_server')"
						:check="show_name_in_sign_sever"
						@toggle="toggleRadio('Show in Sign Server')"
					/>
					<BtRadio
						:label="t('vault_t.can_signing')"
						:check="sign_in_server"
						:disable="!show_name_in_sign_sever"
						@toggle="toggleRadio('Can Signing')"
					/>

					<div class="row items-center justify-between q-my-md">
						<q-btn
							class="bg-grey-11 text-ink-2 text-body1 previous"
							:label="t('vault_t.discard')"
							@click="onPrevious"
						/>
						<q-btn
							class="bg-blue text-body1 next"
							:label="t('create')"
							@click="onFinish"
						/>
					</div>
				</q-step>
			</q-stepper>
		</q-card>
	</q-dialog>
</template>

<script lang="ts">
import { ref, watch } from 'vue';
import { useDialogPluginComponent, useQuasar } from 'quasar';
import { FIELD_DEFS, CryptoTemplate, FieldType } from '@didvault/sdk/src/core';
import ButtonsSlider from '../../../components/ButtonsSlider.vue';
import BottomConfirm from '../../../components/BottomConfirm.vue';
import { app } from '../../../globals';
import { walletService } from '../../../wallet';
import {
	SecertType,
	HostType,
	Secert,
	SecertData
} from '@didvault/sdk/src/core';
import { hexToBytes } from '@didvault/sdk/src/core';
import { AddSecertParams } from '@didvault/sdk/src/core/api';

import { notifyFailed } from '../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';

const stepProgress = {
	'1': '2.1%',
	'2': '50%',
	'3': '100%'
};

export default {
	name: 'CryptoViewAdd',
	props: {},

	emits: [...useDialogPluginComponent.emits],
	components: { ButtonsSlider, BottomConfirm },

	setup() {
		const { CoinType, HDWallet, Mnemonic, PrivateKey, Curve, AnyAddress } =
			walletService.walletCore;

		const CoinTypeDict = {
			Ethereum: CoinType.ethereum,
			BSC: CoinType.ethereum,
			Near: CoinType.near,
			Bitcoin: CoinType.bitcoin
		};

		let step = ref(1);
		let show_name_in_sign_sever = ref(false);
		let sign_in_server = ref(false);

		const { t } = useI18n();

		let walletOption = ref([
			{
				name: 'Ethereum',
				check: false
			},
			{
				name: 'BSC',
				check: false
			},
			{
				name: 'Near',
				check: false
			},
			{
				name: 'Bitcoin',
				check: false
			}
			// {
			//   name: 'OP',
			//   check: true
			// },{
			//   name: 'MATIC',
			//   check: false
			// },{
			//   name: 'AVAX',
			//   check: false
			// },{
			//   name: 'ARB',
			//   check: false
			// },{
			//   name: 'CMT',
			//   check: false
			// }
		]);

		watch(
			() => show_name_in_sign_sever.value,
			(newVal) => {
				if (!newVal) {
					sign_in_server.value = false;
				}
			}
		);

		const { dialogRef, onDialogHide, onDialogOK, onDialogCancel } =
			useDialogPluginComponent();

		async function onOKClick() {
			try {
				onDialogOK();
			} catch (e) {
				console.error(e);
			}
		}

		const $q = useQuasar();
		const options = ref([
			{
				value: 'mnemonic',
				label: t('mnemonic_phrase')
			},
			{
				value: 'privatekey',
				label: t('private_key')
			}
		]);
		const tab = ref('mnemonic');
		const mnemonics = ref('');
		const walletName = ref('');
		const privatekey = ref('');
		const keyStorePassword = ref('');
		const keyStoreFile = ref('');
		const keyStoreName = ref('');
		const onChildChange = (accountType: string) => {
			tab.value = accountType;
		};

		const checkMnemonic = () => {
			if (!mnemonics.value || !Mnemonic.isValid(mnemonics.value)) {
				return false;
			}
			return true;
		};
		const keystorefileUpdate = () => {
			if (keyStoreFile.value == null) {
				keyStoreFile.value = '';
				return;
			}
			if (keyStoreFile.value as unknown as FileList) {
				var input = keyStoreFile.value as unknown as FileList;
				if (input.length > 0) {
					keyStoreName.value = input[0].name;
				} else {
					keyStoreName.value = '';
				}
			}
		};

		const onPrevious = () => {
			if (step.value > 1) {
				step.value--;
			}
		};

		const stringOptions: string[] = Object.keys(CoinTypeDict);
		const tags = ref<string[]>([]);
		const filterOptions = ref<string[]>(stringOptions);

		const onStep2 = () => {
			step.value = 3;

			let tagArr: any = [];
			walletOption.value.map((item) => {
				if (item.check) {
					tagArr.push(item.name);
				}
			});
			tags.value = tagArr;
		};

		const secert: Secert = new Secert({
			secert_type: SecertType.None,
			host_type: HostType.None,
			name: ''
		});

		const onStep1 = () => {
			if (walletName.value.length == 0) {
				notifyFailed(t('errors.please_input_wallet_name'));
				return;
			}
			if (tab.value === options.value[0].value) {
				if (!checkMnemonic()) {
					notifyFailed(t('errors.invalid_mnemonic'));
					return;
				}
			} else if (tab.value === options.value[1].value) {
				if (privatekey.value.length != 64) {
					notifyFailed(t('errors.invalid_privatekey'));
					return;
				}
			} else if (tab.value === options.value[2].value) {
				if (keyStoreFile.value == null) {
					notifyFailed(t('errors.invalid_keystore'));
					return;
				}
				if (keyStoreFile.value as unknown as FileList) {
					var input = keyStoreFile.value as unknown as FileList;
					if (input.length == 0) {
						notifyFailed(t('errors.invalid_keystore'));
						return;
					}
				}
				if (keyStorePassword.value.length == 0) {
					notifyFailed(t('errors.invalid_keyStore_password'));
					return;
				}
			}

			secert.name = walletName.value;
			if (tab.value === options.value[0].value) {
				secert.secert_type = SecertType.Mnemonic;
				secert.data = new SecertData({
					mnemonic: mnemonics.value
				});
			} else if (tab.value === options.value[1].value) {
				if (
					!PrivateKey.isValid(hexToBytes(privatekey.value), Curve.secp256k1)
				) {
					notifyFailed(t('errors.the_private_key_is_invalid_please_check'));
					return;
				}

				secert.secert_type = SecertType.PrivateKey;
				secert.data = new SecertData({
					private_key: privatekey.value,
					curve: 'secp256k1'
				});
			} else if (tab.value === options.value[2].value) {
				// if (keyStoreFile.value as unknown as FileList) {
				//     var input = keyStoreFile.value as unknown as FileList;
				//     var reader = new FileReader();
				//     if (input.length > 0) {
				//         reader.readAsText(input[0]);
				//         reader.onload = async function () {
				//             if (reader.result) {
				//                 //显示文件内容
				//                 const result = reader.result;
				//                 // eslint-disable-next-line @typescript-eslint/no-unused-vars
				//                 Wallet
				//                     .importKeyStore(walletName.value, result as string, keyStorePassword.value).then(() => {
				//                         $q.loading.hide();
				//                     }).catch((e:any) => {
				//                         $q.loading.hide();
				//                     });
				//             }
				//         };
				//     }
				// }
			}
			step.value = 2;
		};

		const onFinish = async () => {
			$q.loading.show();

			try {
				secert.host_type = HostType.None;
				if (show_name_in_sign_sever.value) {
					secert.host_type = HostType.ONLY_SAVE_NAME;
				}

				if (sign_in_server.value) {
					secert.host_type = HostType.SAVE_SECERT;
				}

				if (
					secert.host_type == HostType.ONLY_SAVE_NAME ||
					secert.host_type == HostType.SAVE_SECERT
				) {
					await app.addSecertToSignServer(
						new AddSecertParams({
							secert
						})
					);
				}

				const fields: any = [];

				if (secert.secert_type == SecertType.Mnemonic) {
					const w = HDWallet.createWithMnemonic(secert.data!.mnemonic!, '');
					fields.push({
						name: 'Mnemonic',
						value: mnemonics.value,
						type: FieldType.CryptoeAddress
					});
					for (let chain of tags.value) {
						fields.push({
							name: chain,
							value: w.getAddressForCoin(CoinTypeDict[chain]),
							type: FieldType.CryptoeAddress
						});
					}
				} else if (secert.secert_type == SecertType.PrivateKey) {
					const private_key = PrivateKey.createWithData(
						hexToBytes(secert.data!.private_key!)
					);
					for (let chain of tags.value) {
						fields.push({
							name: chain,
							value: AnyAddress.createWithPublicKey(
								private_key.getPublicKey(CoinTypeDict[chain]),
								CoinTypeDict[chain]
							).description(),
							type: FieldType.CryptoeAddress
						});
					}
				}
				let crypto: CryptoTemplate = {
					name: secert.name,
					tags: tags.value,
					icon: 'vault',
					fields: fields
				};
				onDialogOK(crypto);
			} finally {
				$q.loading.hide();
			}
		};

		const toggleWallet = (name: string) => {
			const w1 = JSON.parse(JSON.stringify(walletOption.value));
			const newArr: {
				name: string;
				check: boolean;
			}[] = [];
			for (let i = 0; i < w1.length; i++) {
				const element = w1[i];
				if (element.name === name) {
					element.check = !element.check;
				}
				newArr.push(element);
			}
			walletOption.value = newArr;
		};

		const toggleRadio = (name: string) => {
			switch (name) {
				case 'Show in Sign Server':
					show_name_in_sign_sever.value = !show_name_in_sign_sever.value;
					break;

				case 'Can Signing':
					sign_in_server.value = !sign_in_server.value;
					break;

				default:
					break;
			}
		};

		return {
			dialogRef,
			step,
			FIELD_DEFS,
			onDialogHide,
			onOKClick,
			onCancelClick: onDialogCancel,
			t,
			options,
			tab,
			mnemonics,
			walletName,
			privatekey,
			keyStorePassword,
			keyStoreFile,
			keyStoreName,
			onChildChange,
			checkMnemonic,
			keystorefileUpdate,
			onStep1,
			onStep2,
			show_name_in_sign_sever,
			sign_in_server,
			onPrevious,
			onFinish,
			tags,
			filterOptions,
			stepProgress,
			walletOption,
			toggleWallet,
			toggleRadio,
			createValue(val: string, done: any) {
				// Calling done(var) when new-value-mode is not set or is "add", or done(var, "add") adds "var" content to the model
				// and it resets the input textbox to empty string
				// ----
				// Calling done(var) when new-value-mode is "add-unique", or done(var, "add-unique") adds "var" content to the model
				// only if is not already set and it resets the input textbox to empty string
				// ----
				// Calling done(var) when new-value-mode is "toggle", or done(var, "toggle") toggles the model with "var" content
				// (adds to model if not already in the model, removes from model if already has it)
				// and it resets the input textbox to empty string
				// ----
				// If "var" content is undefined/null, then it doesn't tampers with the model
				// and only resets the input textbox to empty string

				if (val.length > 0) {
					const modelValue: any = (tags.value || []).slice();

					val
						.split(/[,;|]+/)
						.map((v) => v.trim())
						.filter((v) => v.length > 0)
						.forEach((v) => {
							if (stringOptions.includes(v) === false) {
								stringOptions.push(v);
							}
							if (modelValue.includes(v) === false) {
								modelValue.push(v);
							}
						});

					done(null);
					tags.value = modelValue;
				}
			},

			filterFn(val: string, update: any) {
				update(() => {
					if (val === '') {
						filterOptions.value = stringOptions;
					} else {
						const needle = val.toLowerCase();
						filterOptions.value = stringOptions.filter(
							(v) => v.toLowerCase().indexOf(needle) > -1
						);
					}
				});
			}
		};
	}
};
</script>

<style lang="scss" scoped>
.d-creatVault {
	.q-dialog-plugin {
		width: 500px;
		border-radius: 20px;

		.title {
			width: 100%;
			text-align: center;
		}

		.stepHeader {
			width: 90%;
			height: 10px;
			border-radius: 5px;
			position: relative;
			margin: 0 auto;

			div {
				width: 50%;
				height: 10px;
				border-radius: 5px;
				position: absolute;
				left: 0;
				top: 0;
			}
		}
	}

	.content-bg {
		padding: 0;

		.base-margin {
			margin-top: 16px;
		}

		.input-area-bg {
			@extend .base-margin;
			border-radius: 10px;
			border-width: 1px;
			border-color: $grey-2;
			border-style: solid;
		}

		.input-area {
			margin-top: 0px;
			margin-bottom: 7px;
			min-height: 140px;
			margin-left: 16px;
			margin-right: 16px;
		}

		.section-reminder {
			@extend .base-margin;
			margin-top: 8px;
			text-align: left;
		}

		.section-wallet-name {
			margin-left: 16px;
			margin-right: 16px;
		}
	}

	.step2 {
		width: 100%;
		border: 1px solid $separator;
		border-radius: 10px 10px 4px 4px;
		margin: 0;
		padding-left: 0;

		div {
			padding-left: 0;
		}
	}

	.previous {
		width: 48%;
		height: 56px;
		border-radius: 10px;
		box-shadow: none;
		text-transform: capitalize;

		&::before {
			box-shadow: none;
		}
	}

	.next {
		width: 48%;
		height: 56px;
		border-radius: 10px;
		text-transform: capitalize;

		&::before {
			box-shadow: none;
		}
	}
}
</style>
