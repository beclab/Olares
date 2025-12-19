<template>
	<div class="domain-detail-root">
		<BindTerminusVCContent
			class="bg-background"
			style="height: 100%"
			:has-btn="true"
			:btn-title="btnTitle"
			:btn-status="btnStatusRef"
			@onConfirm="confirmAction"
		>
			<template v-slot:desc>
				<div v-html="desc" />
			</template>
			<template v-slot:content>
				<div class="content q-mt-lg">
					<div
						v-if="
							domain &&
							(domain.status == DOMAIN_STATUS.WAIT_REQUEST_VC ||
								domain.status == DOMAIN_STATUS.WAIT_SUBMIT_VP)
						"
					>
						<div class="module-title text-subtitle3">
							{{ t('Domain name') }}
						</div>
						<div
							class="domain-name text-body2 row items-center q-pl-md q-mt-xs"
						>
							{{ domainName }}
						</div>
					</div>
					<div
						v-else-if="
							domain &&
							(domain.status == DOMAIN_STATUS.BIND ||
								domain.status == DOMAIN_STATUS.BINDING ||
								domain.status == DOMAIN_STATUS.ALLOCATED)
						"
						class="column"
					>
						<div class="module-title text-body3">
							{{ t('select_binding_rules') }}
						</div>
						<q-btn-dropdown
							dropdown-icon="sym_r_expand_more"
							no-caps
							menu-anchor="bottom start"
							:menu-offset="[0, 5]"
							unelevated
							dense
							menu-self="top left"
							class="binding-rule-select q-mt-xs text-ink-2"
						>
							<template v-slot:label>
								<div
									class="row items-center no-wrap q-ml-sm rule-select text-body2"
								>
									<div>{{ method.label }}</div>
								</div>
							</template>
							<q-list>
								<q-item
									v-for="item in methods"
									:key="item.value"
									clickable
									v-close-popup
									@click="method = item"
								>
									<q-item-section>
										<q-item-label class="rule-select text-body2">{{
											item.label
										}}</q-item-label>
									</q-item-section>
								</q-item>
							</q-list>
						</q-btn-dropdown>

						<div
							class="module-title text-body3 q-mt-lg"
							v-if="method.value == 'fixed_email_suffix'"
						>
							{{ t('fixed_email_suffix') }}
						</div>

						<div
							class="q-mt-xs common-border"
							v-if="method.value == 'fixed_email_suffix'"
						>
							<q-input
								v-model="suffixRef"
								borderless
								type="textarea"
								class="q-mx-md"
								:placeholder="
									t(
										'please_enter_your_organization_email_suffix_enter_multiple_separated_by_commas'
									)
								"
							/>
						</div>

						<div
							class="module-title text-body3 q-mt-sm"
							v-else-if="method.value == 'specified_email_address'"
						>
							{{ t('specified_email_address_reminder') }}
						</div>
					</div>
					<div
						v-else-if="
							domain && domain.status == DOMAIN_STATUS.WAIT_TXT_RESOLVE
						"
					>
						<div class="module-title text-body3 q-pt-lg">
							{{ t('name') }}
						</div>
						<div
							class="copy-item text-body2 q-mt-xs row items-center justify-between q-px-md"
							@click="copyFunc(domain.txt.name || '')"
						>
							<div class="content">
								{{ domain.txt.name }}
							</div>
							<q-icon name="sym_r_content_copy" size="20px" class="icon" />
						</div>

						<div class="module-title text-body3 q-mt-lg">
							{{ t('value') }}
						</div>

						<div
							class="copy-item text-body2 q-mt-xs row items-center justify-between q-px-md"
							@click="copyFunc(domain.txt.value || '')"
						>
							<div class="content">
								{{ domain.txt.value }}
							</div>
							<q-icon name="sym_r_content_copy" size="20px" class="icon" />
						</div>
					</div>
					<div
						v-else-if="domain && domain.status == DOMAIN_STATUS.WAIT_NS_RESOLVE"
					>
						<div
							v-for="item in ns"
							:key="item"
							class="copy-item text-body2 q-mt-lg row items-center justify-between q-px-md"
							@click="copyFunc(item || '')"
						>
							<div class="content">
								{{ item }}
							</div>
							<q-icon name="sym_r_content_copy" size="20px" class="icon" />
						</div>
					</div>
				</div>
			</template>
		</BindTerminusVCContent>
		<terminus-title-bar style="position: absolute; top: 0" :translate="true" />
	</div>
</template>

<script lang="ts" setup>
import { useCloudStore, Domain, DOMAIN_STATUS } from '../../../../stores/cloud';
import { ref, onMounted, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useRoute, useRouter } from 'vue-router';
import { Encoder } from '@bytetrade/core';
import {
	getPrivateJWK,
	getDID,
	getEthereumAddress
} from '../../../../did/did-key';
import {
	ItemTemplate,
	cloneItemTemplates,
	Field,
	UserItem,
	VaultType
} from '@didvault/sdk/src/core';
import { getDomainVC } from './domain';
import { app } from '../../../../globals';
import { PrivateJwk, Submission } from '@bytetrade/core';
import { useUserStore } from '../../../../stores/user';
import { submitPresentation } from '../../vc/vcutils';
import BindTerminusVCContent from '../../login/vc/BindTerminusVCContent.vue';
import TerminusTitleBar from '../../../../components/common/TerminusTitleBar.vue';
import { ConfirmButtonStatus } from '../../../../utils/constants';
import { getPlatform } from '@didvault/sdk/src/core';
import {
	notifyFailed,
	notifySuccess
} from '../../../../utils/notifyRedefinedUtil';
import { useI18n } from 'vue-i18n';
import { VCCardInfo } from '../../../../utils/vc';

const { t } = useI18n();
const $q = useQuasar();
const cloudStore = useCloudStore();
const userStore = useUserStore();

const route = useRoute();
const router = useRouter();

const method_value = [
	{
		label: 'Fixed email Suffix',
		value: 'fixed_email_suffix'
	},
	{
		label: 'Specified Email Address',
		value: 'specified_email_address'
	}
];

const methods = ref(method_value);

const method = ref(methods.value[0]);

const domainName = ref('');
const domain = ref<Domain | undefined>();
let result = ref<VCCardInfo | undefined>(undefined);

const suffixRef = ref('');

const ns = ref<string[]>([]);

onMounted(async () => {
	const base: string = route.params.base as string;
	domainName.value = Encoder.bytesToString(Encoder.base64UrlToBytes(base));
	domain.value = await cloudStore.selectDomain(domainName.value);

	if (domain.value && domain.value.status == DOMAIN_STATUS.WAIT_NS_RESOLVE) {
		try {
			let jsonString = domain.value.ns!.replace('[', '');
			jsonString = jsonString.replace(']', '');
			ns.value = jsonString.split(',');
		} catch (error) {
			console.error(error.message);
		}
	}
});

async function requestVC() {
	$q.loading.show();
	try {
		let user: UserItem = userStore.users!.items.get(userStore.current_id!)!;

		let did = await getDID(userStore.current_mnemonic?.mnemonic);
		let privateJWK: PrivateJwk | undefined = await getPrivateJWK(
			userStore.current_mnemonic?.mnemonic
		);
		const owner = await getEthereumAddress(
			userStore.current_mnemonic?.mnemonic
		);

		if (!did) {
			throw new Error(t('errors.get_did_failure'));
		}
		if (!privateJWK) {
			throw new Error(t('errors.get_privatejwk_failure'));
		}

		result.value = await getDomainVC(
			owner,
			did,
			domainName.value,
			user.cloud_id,
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
		template.fields[0].value = domainName.value;
		template.fields[1].value = result.value?.manifest;
		template.fields[2].value = result.value?.verifiable_credential;

		await app.createItem({
			name: 'Domain ' + domainName.value,
			vault,
			fields: template?.fields.map(
				(f) => new Field({ ...f, value: f.value || '' })
			),
			tags: [],
			icon: template?.icon,
			type: VaultType.VC
		});

		const submission: Submission = await submitPresentation(
			'Domain',
			did,
			privateJWK,
			result.value!.verifiable_credential.substring(
				0,
				result.value!.verifiable_credential.length
			),
			owner,
			domainName.value
		);

		if (!submission || submission.status !== 'approved') {
			throw new Error(submission.reason);
		}
		await cloudStore.getDomains();
		domain.value = await cloudStore.selectDomain(domainName.value);
	} catch (error) {
		console.log(error);

		notifyFailed(error.message);
	} finally {
		$q.loading.hide();
	}
}

const onSaveRule = async () => {
	try {
		if (method.value.value == 'fixed_email_suffix') {
			if (!suffixRef.value) {
				notifyFailed(t('errors.please_input_email_suffix'));
				return;
			}
			$q.loading.show();
			await cloudStore.postRule(
				domainName.value,
				method.value.value,
				suffixRef.value
			);
			$q.loading.hide();
			router.go(-2);
			router.push({
				path: '/OrganizationBindVC/' + domainName.value
			});
		} else if (method.value.value == 'specified_email_address') {
			$q.loading.show();
			await cloudStore.postRule(domainName.value, method.value.value, '');
			$q.loading.hide();
			router.go(-2);
		}
	} catch (error) {
		$q.loading.hide();
		notifyFailed(error.message);
	}
};

const btnStatusRef = computed(function () {
	if (!domain.value) {
		return ConfirmButtonStatus.normal;
	}
	if (
		domain.value.status == DOMAIN_STATUS.WAIT_TXT_RESOLVE ||
		domain.value.status == DOMAIN_STATUS.WAIT_NS_RESOLVE
	) {
		return ConfirmButtonStatus.normal;
	}
	if (
		domain.value.status == DOMAIN_STATUS.WAIT_REQUEST_VC ||
		domain.value.status == DOMAIN_STATUS.WAIT_SUBMIT_VP
	) {
		return ConfirmButtonStatus.normal;
	}

	if (
		domain.value.status == DOMAIN_STATUS.BIND ||
		domain.value.status == DOMAIN_STATUS.BINDING ||
		domain.value.status == DOMAIN_STATUS.ALLOCATED
	) {
		return (method.value.value == 'fixed_email_suffix' &&
			suffixRef.value.length > 0) ||
			method.value.value != 'fixed_email_suffix'
			? ConfirmButtonStatus.normal
			: ConfirmButtonStatus.disable;
	}

	return ConfirmButtonStatus.normal;
});

const btnTitle = computed(function () {
	if (!domain.value) {
		return t('continue');
	}
	if (
		domain.value.status == DOMAIN_STATUS.WAIT_TXT_RESOLVE ||
		domain.value.status == DOMAIN_STATUS.WAIT_NS_RESOLVE
	) {
		return t('i_got_it');
	}

	if (
		domain.value.status == DOMAIN_STATUS.WAIT_REQUEST_VC ||
		domain.value.status == DOMAIN_STATUS.WAIT_SUBMIT_VP
	) {
		return t('confirm');
	}

	if (
		domain.value.status == DOMAIN_STATUS.BIND ||
		domain.value.status == DOMAIN_STATUS.BINDING ||
		domain.value.status == DOMAIN_STATUS.ALLOCATED
	) {
		return t('submit');
	}

	return t('back');
});

const desc = computed(function () {
	if (!domain.value) {
		return '';
	}
	if (domain.value.status == DOMAIN_STATUS.WAIT_TXT_RESOLVE) {
		// return 'I got it'
		return t('confirm_bind_domain_text_record_desc');
	}

	if (domain.value.status == DOMAIN_STATUS.WAIT_NS_RESOLVE) {
		return t('confirm_bind_domain_cname_record_desc', {
			domain: domain.value.domain
		});
	}

	if (
		domain.value.status == DOMAIN_STATUS.WAIT_REQUEST_VC ||
		domain.value.status == DOMAIN_STATUS.WAIT_SUBMIT_VP
	) {
		return t('confirm_bind_domain_name_desc');
	}

	if (
		domain.value.status == DOMAIN_STATUS.BIND ||
		domain.value.status == DOMAIN_STATUS.BINDING ||
		domain.value.status == DOMAIN_STATUS.ALLOCATED
	) {
		return t('confirm_bind_domain_bingding_desc');
	}

	return t('confirm_bind_domain_common_desc');
});

const confirmAction = () => {
	if (!domain.value) {
		return '';
	}
	if (
		domain.value.status == DOMAIN_STATUS.WAIT_TXT_RESOLVE ||
		domain.value.status == DOMAIN_STATUS.WAIT_NS_RESOLVE
	) {
		router.go(-1);
		return;
	}

	if (
		domain.value.status == DOMAIN_STATUS.WAIT_REQUEST_VC ||
		domain.value.status == DOMAIN_STATUS.WAIT_SUBMIT_VP
	) {
		requestVC();
		return;
	}

	if (
		domain.value.status == DOMAIN_STATUS.BIND ||
		domain.value.status == DOMAIN_STATUS.BINDING ||
		domain.value.status == DOMAIN_STATUS.ALLOCATED
	) {
		onSaveRule();
		return;
	}

	router.go(-1);
};

const copyFunc = async (content: string) => {
	try {
		await getPlatform().setClipboard(content);
		notifySuccess(t('copy_success'));
	} catch (error) {
		console.error(error.message);
	}
};
</script>

<style lang="scss" scoped>
.domain-detail-root {
	width: 100%;
	height: 100%;
	position: relative;

	.module-title {
		text-align: left;
		color: $grey-5;
	}

	.domain-name {
		border: 1px solid $separator;
		width: 100%;
		height: 36px;
		border-radius: 8px;
		background: $grey-1;
		color: $ink-1;
	}

	.copy-item {
		border: 1px solid $separator;
		width: 100%;
		height: 36px;
		border-radius: 8px;

		color: $ink-1;

		.content {
			max-width: 80%;
			text-overflow: ellipsis;
			white-space: nowrap;
			overflow: hidden;
		}

		.icon {
			color: $blue-4;
		}
	}

	.binding-rule-select {
		height: 36px;
		border-radius: 8px;
		border: 1px solid $separator;
		margin-right: 4px;

		.rule-select {
			color: $ink-1;
			width: 100%;
		}
	}

	::placeholder {
		color: $grey-5;
		opacity: 1;
		font-size: map-get($map: $body2, $key: size);
		font-weight: map-get($map: $body2, $key: weight);
		line-height: map-get($map: $body2, $key: line-height);
		letter-spacing: map-get($map: $body2, $key: letter-spacing);
	}

	::-ms-input-placeholder {
		color: $grey-5;
	}

	.common-border {
		border: 1px solid $separator;
		border-radius: 8px;
	}
}
</style>
