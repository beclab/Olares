<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('third_party_domain')"
		:skip="false"
		size="medium"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:okDisabled="!confirmEnable"
		:platform="deviceStore.platform"
		@onSubmit="onOKClick"
	>
		<div>
			<terminus-edit
				:label="t('third_party_domain')"
				v-model="data"
				:is-error="data.length > 0 && !isValidDomain(data)"
				:error-message="t('Invalid Domain')"
			/>

			<terminus-textarea-edit
				class="q-mt-md"
				v-if="
					reverseProxyMode == ReverseProxyMode.OlaresTunnel ||
					reverseProxyMode == ReverseProxyMode.SelfBuiltFrp
				"
				:label="t('Upload HTTPS Certificate')"
				v-model="cert"
			/>

			<terminus-textarea-edit
				class="q-mt-md"
				v-if="
					reverseProxyMode == ReverseProxyMode.OlaresTunnel ||
					reverseProxyMode == ReverseProxyMode.SelfBuiltFrp
				"
				:label="t('Upload HTTPS Private Key')"
				v-model="key"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { computed, ref } from 'vue';
import { useDeviceStore } from 'src/stores/device';
import { ReverseProxyMode } from 'src/constant/index';
import TerminusEdit from '../../../base/TerminusEdit.vue';
import TerminusTextareaEdit from '../../../base/TerminusTextareaEdit.vue';

const props = defineProps({
	reverseProxyMode: {
		type: Number,
		required: false,
		default: ReverseProxyMode.NoNeed
	},
	domain: {
		type: String,
		required: false,
		default: ''
	}
});

const data = ref(props.domain);

const cert = ref('');

const key = ref('');

const { t } = useI18n();

const CustomRef = ref();
const deviceStore = useDeviceStore();
async function onOKClick() {
	CustomRef.value.onDialogOK({
		data: data.value,
		cert: cert.value,
		key: key.value
	});
}

const confirmEnable = computed(() => {
	if (
		props.reverseProxyMode == ReverseProxyMode.OlaresTunnel ||
		props.reverseProxyMode == ReverseProxyMode.SelfBuiltFrp
	) {
		return (
			data.value &&
			isValidDomain(data.value) &&
			cert.value &&
			cert.value.length > 0 &&
			key.value &&
			key.value.length > 0
		);
	}
	return data.value && data.value.length > 0;
});

function isValidDomain(domain: string) {
	const domainRegex = /^(?!-)[A-Za-z0-9-]{1,63}(?<!-)(\.[A-Za-z]{2,63})+$/;
	if (domain.length > 253) {
		return false;
	}
	return domainRegex.test(domain);
}

// const httpCertRule = () => {
// 	let certAvailable = isHttpsCertificate(
// 		cert.value,
// 		'-----BEGIN CERTIFICATE-----',
// 		'-----END CERTIFICATE-----'
// 	);
// 	if (!certAvailable) {
// 		return t('Invalid HTTPS certificate');
// 	}
// 	return '';
// };

// const httpKeyRule = () => {
// 	let certAvailable = isHttpsCertificate(
// 		key.value,
// 		'-----BEGIN RSA PRIVATE KEY-----',
// 		'-----END RSA PRIVATE KEY-----'
// 	);
// 	if (!certAvailable) {
// 		return t('Invalid HTTPS Private Key');
// 	}
// 	return '';
// };

// function isHttpsCertificate(
// 	certString: string,
// 	beginCert: string,
// 	endCert: string
// ) {
// 	if (typeof certString !== 'string') {
// 		return false;
// 	}
// 	if (!certString.startsWith(beginCert) || !certString.endsWith(endCert)) {
// 		return false;
// 	}

// 	const base64Content = certString
// 		.replaceAll(beginCert, '')
// 		.replaceAll(endCert, '')
// 		.replace(/\r?\n|\r/g, '')
// 		.trim();

// 	const base64Regex = /^[A-Za-z0-9+/=]+$/;
// 	if (!base64Regex.test(base64Content)) {
// 		return false;
// 	}

// 	return true;
// }
</script>
