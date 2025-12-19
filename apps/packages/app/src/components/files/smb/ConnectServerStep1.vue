<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.connect_to_server')"
		:skip="false"
		:okLoading="loading ? t('loading') : ''"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		:persistent="true"
		size="medium"
		@onSubmit="submit"
		@onCancel="onCancel"
	>
		<div class="dialog-desc">
			<div>
				<div class="form-item-key text-body1 text-ink-3 q-mb-xs">
					{{ t('files.dir_address') }}
				</div>
				<div class="form-item-value row item-center justify-between">
					<q-input
						dense
						borderless
						no-error-icon
						v-model="connectUrl"
						:placeholder="t('files.server_address_placeholder')"
						class="form-item-input text-ink-2 text-body1"
						:rules="[
							(val) =>
								(val && val.length > 0) ||
								t('files.server_address_placeholder'),
							(val) => /^\/\//.test(val) || t('files.server_address_rules')
						]"
					>
					</q-input>
				</div>
			</div>

			<div class="q-mt-lg">
				<div class="form-item-key text-body1 text-ink-3 q-mb-xs">
					{{ t('files.Favorite Servers') }}
				</div>
				<div class="favorite-list">
					<BtScrollArea style="height: 100%; width: 100%">
						<div
							class="favorite-item text-ink-2 text-body1 q-px-md"
							:class="{
								'favorite-item-active': favoriteActive === item.url
							}"
							v-for="item in server_address"
							:key="item.value"
							@click="handleActive(item)"
						>
							{{ item.url }}
						</div>
					</BtScrollArea>
				</div>
			</div>
		</div>
		<template v-slot:footerMore>
			<div class="footerMore">
				<ConnectServerFooterEdit
					:show-add="showAdd"
					:show-remove="showRemove"
					@handle-add="saveFavorite"
					@handle-remove="removeFavorite"
				/>
			</div>
		</template>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
import { ref, computed, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { CommonFetch } from '../../../api';
import { useFilesStore, SmbMountType } from './../../../stores/files';
import ConnectServerFooterEdit from './ConnectServerFooterEdit.vue';
import { files } from 'jszip';

const props = defineProps({
	origin_id: {
		type: Number,
		required: true
	}
});

const { t } = useI18n();

const filesStore = useFilesStore();
const route = useRoute();

const loading = ref(false);
const connectUrl = ref();
const favoriteActive = ref();

const CustomRef = ref();

const server_address = ref<any[]>([]);

const showAdd = computed(() => {
	if (connectUrl.value) {
		const hasFavorite = server_address.value.find(
			(item) => item.url === connectUrl.value
		);

		if (hasFavorite) {
			return false;
		} else {
			return true;
		}
	} else {
		return false;
	}
});
const showRemove = computed(() => {
	if (favoriteActive.value) {
		return true;
	} else {
		return false;
	}
});

const saveFavorite = async () => {
	const params = [{ url: connectUrl.value }];
	const extend = filesStore.currentNode[props.origin_id].name;
	await CommonFetch.put('/api/smb_history/' + extend + '/', params);
	fetchData(filesStore.currentNode[props.origin_id].name);
};

const removeFavorite = async () => {
	const params = [{ url: favoriteActive.value }];
	const extend = filesStore.currentNode[props.origin_id].name;
	await CommonFetch.delete('/api/smb_history/' + extend + '/', {
		data: params
	});
	fetchData(filesStore.currentNode[props.origin_id].name);

	favoriteActive.value = null;
};

const handleActive = (item: any) => {
	favoriteActive.value = item.url;
	connectUrl.value = item.url;
};

const mountSmb = async (connectData: SmbMountType) => {
	loading.value = true;
	try {
		const res = await filesStore.mountSmbInExternal(connectData);
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
		CustomRef.value.onDialogOK({
			connectUrl: connectUrl.value,
			hasFavorite: true
		});
		loading.value = false;
	}
};

const submit = async () => {
	const hasFavorite = server_address.value.find(
		(item) => item.url === connectUrl.value
	);

	if (hasFavorite && hasFavorite.username && hasFavorite.password) {
		await mountSmb(hasFavorite);
		return false;
	}

	if (hasFavorite && (!hasFavorite.username || !hasFavorite.password)) {
		CustomRef.value.onDialogOK({
			connectUrl: connectUrl.value,
			hasFavorite: true
		});
		return false;
	}

	CustomRef.value.onDialogOK({
		connectUrl: connectUrl.value,
		hasFavorite: false
	});
};

const onCancel = () => {
	// store.closeHovers();
	// onDialogCancel();
};

const fetchData = async (extend: string) => {
	const data = await CommonFetch.get('/api/smb_history/' + extend + '/', {});
	server_address.value = data || [];
};

onMounted(() => {
	fetchData(filesStore.currentNode[props.origin_id].name);
});
</script>

<style lang="scss" scoped>
.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.form-item-input {
		flex: 1;
		border: 1px solid $input-stroke;
		border-radius: 8px;
		padding: 0 10px;
		box-sizing: border-box;
	}

	.form-item {
		&.margin-top {
			margin-top: 30px;
		}
		.form-item-key {
			width: 100%;
		}
		.form-item-value {
			width: 100%;
		}
	}

	.favorite-list {
		width: 100%;
		height: 200px;
		border: 1px solid $input-stroke;
		border-radius: 8px;
		.favorite-item {
			height: 40px;
			line-height: 40px;
			box-sizing: border-box;
			border-bottom: 1px solid $input-stroke;
			cursor: pointer;
			&.favorite-item-active {
				background-color: rgb(0, 0, 0, 0.03);
			}
		}
		.favorite-item:last-child {
			border-bottom: none;
		}
	}
}

.footerMore {
	width: 100px;
	position: absolute;
	left: 20px;
	border-radius: 8px;
	font-weight: 500;
	font-size: 16px;
	padding: 8px 0;
	line-height: 24px;
}
</style>
