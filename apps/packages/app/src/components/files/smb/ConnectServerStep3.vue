<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.connect_to_server')"
		:skip="false"
		:okLoading="loading ? t('loading') : ''"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:persistent="true"
		@onSubmit="submit"
		@onCancel="onCancel"
	>
		<div class="dialog-desc row items-center justify-between">
			<div class="connect-icon row items-center justify-center q-mr-md">
				<q-icon name="sym_r_language" size="24px" color="white" />
			</div>
			<div class="connect-content">
				<div class="text-ink-1 text-body2">Connecting</div>
				<div class="text-ink-1 text-body2">{{ smb_url }}</div>
			</div>
		</div>

		<div class="dialog-desc q-mt-lg">
			<div class="form-item row">
				<div class="form-item-key text-body1 text-ink-3">
					{{ t('files.name') }}:
				</div>
				<div class="form-item-value">
					<q-input
						dense
						borderless
						no-error-icon
						v-model="connectData.username"
						:placeholder="t('files.server_username_placeholder')"
						class="form-item-input text-ink-2"
						:rules="[
							(val) =>
								(val && val.length > 0) ||
								t('files.server_username_placeholder')
						]"
					>
					</q-input>
				</div>
			</div>

			<div class="form-item row q-mt-lg">
				<div class="form-item-key text-body1 text-ink-3">
					{{ t('files.file_password') }}:
				</div>
				<div class="form-item-value">
					<q-input
						dense
						borderless
						no-error-icon
						v-model="connectData.password"
						type="password"
						:placeholder="t('files.server_password_placeholder')"
						class="form-item-input text-ink-2"
						:rules="[
							(val) =>
								(val && val.length > 0) ||
								t('files.server_password_placeholder')
						]"
					>
					</q-input>
				</div>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, reactive } from 'vue';
import { useRoute } from 'vue-router';
import { CommonFetch } from '../../../api';
import { useDataStore } from '../../../stores/data';
import { useFilesStore } from '../../../stores/files';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	},
	smb_url: {
		type: String,
		required: true
	},
	hasFavorite: {
		type: Boolean,
		default: false
	}
});

const store = useDataStore();
const { t } = useI18n();

const CustomRef = ref();

const filesStore = useFilesStore();
const route = useRoute();

const loading = ref(false);
const connectData = reactive({
	url: props.smb_url,
	username: '',
	password: ''
});

const saveSmbHistory = async (extend: string) => {
	await CommonFetch.put('/api/smb_history/' + extend + '/', [connectData]);
};

const submit = async () => {
	loading.value = true;

	try {
		const res = await filesStore.mountSmbInExternal(connectData);

		if (props.hasFavorite) {
			saveSmbHistory(filesStore.currentNode[props.origin_id].name);
		}

		if (res.code === 300) {
			return CustomRef.value.onDialogOK({
				connectData,
				paths: res.data
			});
		}

		CustomRef.value.onDialogOK();
		filesStore.setBrowserUrl(
			route.fullPath,
			filesStore.activeMenu(props.origin_id).driveType
		);
		loading.value = false;
	} catch (error) {
		console.log('error');
		loading.value = false;
	}
};

const onCancel = () => {
	store.closeHovers();
};
</script>

<style lang="scss" scoped>
.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.connect-icon {
		width: 40px;
		height: 40px;
		border-radius: 10px;
		background-color: $light-blue-default;
	}
	.connect-content {
		flex: 1;
	}

	.form-item-input {
		border: 1px solid $input-stroke;
		border-radius: 8px;
		padding: 0 10px;
		box-sizing: border-box;
	}

	.form-item {
		.form-item-key {
			width: 100%;
			height: 40px;
			line-height: 40px;
		}
		.form-item-value {
			width: 100%;
		}
	}
}
</style>
