<template>
	<bt-custom-dialog
		ref="dialogRef"
		:title="title"
		:ok="okText"
		:cancel="t('cancel')"
		size="medium"
		:platform="$q.platform.is.mobile ? 'mobile' : 'web'"
		:okDisabled="onDisabled"
		@onSubmit="unlock"
	>
		<div class="archive-password-dialog">
			<div class="archive-password-banner row items-start no-wrap">
				<q-icon name="sym_r_lock" size="24px" class="q-mr-sm" />
				<div class="text-body1 text-ink-1">
					{{ t('files.archive.encrypted_archive_tip', { name: archiveName }) }}
				</div>
			</div>

			<terminus-edit
				v-model="password"
				:show-password-img="true"
				title-class="archive-name-title"
				:label="t('files.archive.password')"
				:hint-text="t('files.archive.enter_archive_password')"
				:inputHeight="40"
				:is-error="
					errorPassword.length > 0 &&
					password == errorPassword &&
					errorMessage.length > 0
				"
				:error-message="errorMessage"
				@keyup.enter="unlock"
			/>
		</div>
	</bt-custom-dialog>
</template>

<script setup lang="ts">
import { computed, ref } from 'vue';
import { useQuasar } from 'quasar';
import { useI18n } from 'vue-i18n';
import TerminusEdit from '../../common/TerminusEdit.vue';
import { i18n } from 'src/boot/i18n';

const props = defineProps({
	archiveName: {
		type: String,
		default: ''
	},
	errorMessage: {
		type: String,
		default: ''
	},
	title: {
		type: String,
		default: ''
	},
	okText: {
		type: String,
		default: i18n.global.t('confirm')
	},
	errorPassword: {
		type: String,
		default: ''
	}
});

const $q = useQuasar();
const { t } = useI18n();
const dialogRef = ref();
const password = ref(props.errorPassword);
const archiveName = props.archiveName || t('files.archive.this_archive');
const onDisabled = computed(() => password.value.length === 0);

const unlock = () => {
	if (!password.value) {
		return;
	}
	dialogRef.value.onDialogOK(password.value);
};
</script>

<style scoped lang="scss">
.archive-password-dialog {
	width: 520px;
	display: flex;
	flex-direction: column;
	gap: 20px;
}

@media (max-width: 599px) {
	.archive-password-dialog {
		width: calc(100vw - 40px);
		max-width: 100%;
	}
}

.archive-password-banner {
	width: 100%;
	// min-height: 72px;
	padding: 12px;
	border-radius: 8px;
	background: $orange-alpha;
}
</style>

<style lang="scss">
.archive-name-title {
	color: $ink-3;
}
</style>
