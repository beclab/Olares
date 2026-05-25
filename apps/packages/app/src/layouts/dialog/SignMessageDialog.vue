<template>
	<q-dialog
		ref="root"
		@hide="onDialogHide"
		maximized
		transition-show="slide-up"
		transition-hide="slide-down"
	>
		<q-card class="q-dialog-plugin column root">
			<div
				class="header row justify-between items-center"
				:style="
					$q.platform.is.nativeMobile && $q.platform.is.android
						? 'margin-top:30px;'
						: ''
				"
			>
				<div class="row justify-start items-center">
					<div
						class="icon-container column justify-center items-center"
						@click="onCancelClick"
					>
						<q-icon name="sym_r_close" size="24px" />
					</div>
				</div>
			</div>

			<div class="sign-content-container column items-center">
				<bt-scroll-area class="sign-content-container__content">
					<div
						class="column items-center full-width"
						style="padding-bottom: 60px"
					>
						<div
							class="row items-center justify-center"
							style="width: 124px; height: 124px; margin-top: 64px"
						>
							<q-img
								:src="defaultMessageInfo(body).imgPath"
								width="90px"
								height="90px"
								noSpinner
								style="border-radius: 50%"
							/>
						</div>

						<div class="text-ink-1 text-h5 title">
							{{ defaultMessageInfo(body).title }}
						</div>

						<div class="text-ink-2 text-body3 content">
							{{ defaultMessageInfo(body).content }}
						</div>
						<div
							class="full-width user-info q-mt-lg q-py-lg column items-center justify-center"
							v-if="body.event == 'login_cloud'"
						>
							<div class="users">
								<TerminusAvatar :info="userStore.terminusInfo()" :size="40" />
							</div>
							<div class="name text-h6 text-ink-1">
								{{ userStore.current_user?.local_name }}
							</div>
							<div class="did text-ink-3 text-body3">
								{{ userStore.current_user?.name }}
							</div>
						</div>
					</div>
				</bt-scroll-area>

				<q-btn
					class="confirm row items-center justify-center"
					@click="onOKClick"
					flat
					no-caps
					:disable="!confirmEnable"
				>
					<q-spinner-dots color="text-ink-2" v-if="!confirmEnable" />
					<div v-else class="text-grey-10">{{ t('confirm') }}</div>
				</q-btn>
				<q-btn class="cancel" flat no-caps @click="onCancelClick">
					<div>{{ t('cancel') }}</div>
				</q-btn>
			</div>
		</q-card>
	</q-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted, PropType, onUnmounted } from 'vue';
import { useQuasar } from 'quasar';
import { app } from '../../globals';
import { useUserStore } from '../../stores/user';
import { MessageBody } from '@bytetrade/core';
import { getPrivateJWK, getDID, getEthereumAddress } from '../../did/did-key';
import {
	ItemTemplate,
	cloneItemTemplates,
	Field,
	UserItem,
	VaultType,
	VaultItem,
	FieldType
} from '@didvault/sdk/src/core';
import {
	signJWS,
	requestVC,
	signStatement,
	mnemonicToKey,
	defaultDriverPath
} from './sign';
import { PrivateJwk, Submission } from '@bytetrade/core';
import { submitPresentation } from '../../pages/Mobile/vc/vcutils';
import { busOn, busOff } from '../../utils/bus';
import { notifySuccess, notifyFailed } from '../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { i18n } from '../../boot/i18n';
import { generateStringEllipsis } from '../../utils/utils';
import { getRequireImage } from '../../utils/imageUtils';
import { axiosInstanceProxy } from 'src/platform/httpProxy';
import { getNativeAppPlatform } from 'src/application/platform';

const props = defineProps({
	body: {
		type: Object as PropType<MessageBody>,
		required: true
	}
});
const $q = useQuasar();

const sign_id = props.body.message?.id;

const emit = defineEmits(['hide']);
const root = ref<any>(null);

const confirmEnable = ref(true);

const userStore = useUserStore();

const { t } = useI18n();

const onDialogHide = () => {
	emit('hide');
};

onMounted(async () => {
	busOn('cancel_sign', (data) => {
		if (data == sign_id) {
			emit('hide');
		}
	});
});

onUnmounted(() => {
	busOff('cancel_sign');
});

const onOKClick = async () => {
	confirmEnable.value = false;

	try {
		let user: UserItem = userStore.users!.items.get(userStore.current_id!)!;
		let mneminicItem = userStore.current_mnemonic;
		if (!mneminicItem) {
			throw new Error(t('errors.mnemonics_is_not_valid'));
		}
		let did = await getDID(mneminicItem.mnemonic);
		let privateJWK: PrivateJwk | undefined = await getPrivateJWK(
			mneminicItem.mnemonic
		);
		const owner = await getEthereumAddress(mneminicItem.mnemonic);

		if (!did) {
			throw new Error(t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(t('errors.get_privatejwk_failure'));
		}
		const callback_url = props.body.message?.sign?.callback_url;
		let headers: any = {
			'Content-Type': 'application/json'
		};
		let params: any = {};

		if (callback_url && userStore.current_user?.name) {
			const callbackUrlObj = new URL(callback_url);
			if (
				callbackUrlObj.host.includes(
					userStore.current_user.name.replace('@', '.')
				)
			) {
				headers['X-Authorization'] = userStore.current_user?.access_token || '';
				params = {
					hideCookie: true
				};
			}
		}

		const axoisProxy = axiosInstanceProxy(
			{
				headers,
				params
			},
			false
		);

		const appPlatform = getNativeAppPlatform();

		const item = appPlatform.signProtocolList.find((item) => {
			const result = item.precheck(props.body);
			return result;
		});

		if (item && item.signAction) {
			const result = await item.signAction(props.body, {
				did,
				privateJWK
			});
			await axoisProxy.post(result.callback_url, result.postData);
		} else {
			if (props.body.message?.sign?.sign_vc) {
				await requestVCVP(
					user,
					did,
					privateJWK,
					owner,
					props.body.message?.sign?.sign_vc.type,
					props.body.message?.sign?.sign_vc.name,
					props.body.message?.sign?.sign_vc.request_path,
					props.body.message?.sign?.sign_vc.data
				);
				const url = props.body.message?.sign?.callback_url;
				const postData = {
					id: props.body.message?.id
				};
				await axoisProxy.post(url!, postData);
			} else {
				let eth721Sign = '';

				if (props.body.message?.sign?.sign_eth) {
					const ownerKey = mnemonicToKey(
						mneminicItem.mnemonic,
						defaultDriverPath(0)
					);
					eth721Sign = await signStatement(
						props.body.message?.sign?.sign_eth.domain,
						props.body.message?.sign?.sign_eth.types,
						props.body.message?.sign?.sign_eth.data,
						props.body.message?.sign?.sign_eth.primaryType,
						ownerKey
					);
				}

				let body = { ...props.body.message?.sign?.sign_body };
				if (eth721Sign) {
					body['eth721_sign'] = eth721Sign;
				}

				const jws = await signJWS(did, body, privateJWK);

				const url = props.body.message?.sign?.callback_url;
				const postData = {
					id: props.body.message?.id,
					jws,
					did,
					...body
				};

				await axoisProxy.post(url!, postData);
			}
		}

		if (item && item.afterSign) {
			await item.afterSign(props.body);
		}

		notifySuccess(t('sign_success'));
		emit('hide');
	} catch (e) {
		notifyFailed(e.message);
	} finally {
		confirmEnable.value = true;
		$q.loading.hide();
	}
};

const onCancelClick = () => {
	emit('hide');
};

async function requestVCVP(
	user: UserItem,
	did: string,
	privateJWK: PrivateJwk,
	owner: string,
	vc_type: string,
	vc_name: string,
	vc_request_path: string,
	vc_sign_data: any
) {
	const vcResult = await requestVC(
		did,
		vc_type,
		vc_request_path,
		vc_sign_data,
		privateJWK
	);

	const vault = app.mainVault;
	if (!vault) {
		throw new Error(t('errors.main_vault_is_null'));
	}

	const template: ItemTemplate | undefined = cloneItemTemplates().find(
		(template) => template.id === 'vc'
	);
	if (!template) {
		throw new Error(t('errors.template_is_null'));
	}
	template.fields[0].value = vc_name;
	template.fields[1].value = vcResult.manifest;
	template.fields[2].value = vcResult.verifiable_credential;

	await app.createItem({
		name: vc_name,
		vault,
		fields: template?.fields.map(
			(f) => new Field({ ...f, value: f.value || '' })
		),
		tags: [],
		icon: template?.icon,
		type: VaultType.VC
	});

	const submission: Submission = await submitPresentation(
		vc_type,
		did,
		privateJWK,
		vcResult.verifiable_credential.substring(
			0,
			vcResult.verifiable_credential.length
		),
		owner,
		null
	);

	if (!submission || submission.status !== 'approved') {
		throw new Error(submission.reason);
	}
	return '';
}

const defaultMessageInfo = (messageBody: MessageBody) => {
	const usestore = useUserStore();
	let title = messageBody.notification?.title || '';
	let content = messageBody.notification?.body || '';
	let imgPath = messageBody.app?.icon || '';
	if (messageBody.event == 'login_cloud') {
		let contentStr = messageBody.notification?.body;
		const didReplace = usestore.current_user?.name
			? usestore.current_user?.name
			: generateStringEllipsis(usestore.current_user?.id || '', 17);

		if (contentStr) {
			contentStr = contentStr.replace(/'did'/g, didReplace);
		} else {
			contentStr = i18n.global.t('login_olares_space_desc', {
				did: didReplace
			});
		}

		title =
			messageBody.notification?.title || i18n.global.t('login_olares_space');
		content = contentStr;
		imgPath = getRequireImage('cloud/login/cloud-logo.png');
	}
	return {
		title,
		content,
		imgPath
	};
};
</script>

<style lang="scss" scoped>
.root {
	border-radius: 10px;
	padding: 10px 20px;
	background: $background-2;
	.header {
		height: 40px;
		width: 100%;
		text-align: center;
		padding: 0;
		position: relative;

		.icon-container {
			width: 32px;
			height: 32px;
		}
	}

	.sign-content-container {
		background: $background-1;
		height: calc(100% - 81px);
		border-radius: 12px;
		width: 100%;

		&__content {
			width: 100%;
			height: calc(100% - 120px);

			.title {
				text-align: center;
			}
			.content {
				text-align: center;
				width: calc(100% - 64px);
				word-break: break-all;
				overflow-wrap: break-word;
				white-space: pre-wrap;
				overflow: hidden;
			}

			.user-info {
				// height: 128px;
				border-radius: 12px;
				border: 1px solid $separator;

				.users {
					width: 40px;
					height: 40px;
					border-radius: 20px;
					overflow: hidden;
					margin-left: 10px;
				}

				// .info {
				// 	flex: 1;
				// 	overflow: hidden;
				// }
				.name {
					color: $ink-1;
					// height: 24px;
					word-break: break-all;
					overflow-wrap: break-word;
					white-space: pre-wrap;
					overflow: hidden;
					max-width: 80%;
				}

				.did {
					color: $ink-2;
					// height: 16px;
					word-break: break-all;
					overflow-wrap: break-word;
					white-space: pre-wrap;
					// overflow: hidden;
					max-width: 80%;
				}
			}
		}

		.confirm {
			width: 207px;
			height: 48px;
			background: $yellow;
			border-radius: 8px;

			&:before {
				box-shadow: none;
			}
		}

		.cancel {
			width: 46%;
			height: 48px;
			padding-top: 10px;
			border-radius: 10px;
			box-shadow: none;
			color: $blue-4;

			&:before {
				box-shadow: none;
			}
		}
	}

	.password {
		width: 100%;
		margin-top: 10px;
		border: 1px solid $separator;
		border-radius: 10px;
		padding: 0 10px;
	}

	.button {
		width: 100%;
		margin-top: 20px;
	}
}
</style>
