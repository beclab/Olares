<template>
	<bt-custom-dialog
		ref="CustomRef"
		:title="t('files.select_mount_dir')"
		:skip="false"
		:okLoading="loading ? t('loading') : ''"
		:okDisabled="pathActive ? false : true"
		:ok="t('confirm')"
		:cancel="t('cancel')"
		size="medium"
		:persistent="true"
		@onSubmit="submit"
	>
		<div class="dialog-desc">
			<div class="form-item-key text-ink-3 q-mb-xs">
				{{ t('files.dir_address') }}: {{ connectData.url }}
			</div>
			<div class="path-list">
				<BtScrollArea style="height: 100%; width: 100%">
					<div
						class="path-item text-ink-2 text-body2 q-px-md row items-center justify-between"
						:class="{
							'path-item-active': pathActive === item.path,
							'item-scroll': props.paths.length >= 5
						}"
						v-for="item in props.paths"
						:key="item.path"
						@click="handleActive(item)"
					>
						<span
							class="path-name text-body1"
							:class="{ 'path-disabled': item.mounted }"
						>
							{{ item.dir }}
						</span>
						<span
							v-if="item.mounted"
							class="text-light-blue-default text-body2"
							:class="{ 'mounted-disabled': item.mounted }"
						>
							{{ t('files.mounted') }}
						</span>
					</div>
				</BtScrollArea>
			</div>
		</div>
	</bt-custom-dialog>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';
import { useFilesStore, SmbMountType } from './../../../stores/files';

interface Props {
	origin_id: number;
	connectData: SmbMountType;
	paths: any[];
}
const props = withDefaults(defineProps<Props>(), {});

const filesStore = useFilesStore();

const { t } = useI18n();
const route = useRoute();

const loading = ref(false);
const pathActive = ref();
const CustomRef = ref();

const handleActive = (item: any) => {
	if (item.mounted) return;
	pathActive.value = item.path;
};

const submit = async () => {
	if (pathActive.value) {
		loading.value = true;
		try {
			await filesStore.mountSmbInExternal({
				...props.connectData,
				url: pathActive.value
			});
			loading.value = false;
			CustomRef.value.onDialogOK();
			filesStore.setBrowserUrl(
				route.fullPath,
				filesStore.activeMenu(props.origin_id).driveType
			);
		} catch (z) {
			loading.value = false;
		}
	}
};
</script>

<style lang="scss" scoped>
.dialog-desc {
	width: 100%;
	padding: 0 0px;

	.path-list {
		width: 100%;
		height: 200px;
		border: 1px solid $input-stroke;
		border-radius: 8px;
		.path-item {
			width: 100%;
			height: 40px;
			line-height: 40px;
			box-sizing: border-box;
			border-bottom: 1px solid $input-stroke;
			cursor: pointer;

			.path-name {
				display: inline-block;
				width: 428px;
				overflow: hidden;
				text-overflow: ellipsis;
				white-space: nowrap;
				&.path-disabled {
					opacity: 0.5;
					cursor: default;
				}
			}
			.mounted-disabled {
				cursor: default;
			}
			&.path-item-active {
				background-color: rgb(0, 0, 0, 0.03);
			}
		}
		.item-scroll.path-item:last-child {
			border-bottom: none;
		}
	}
}
</style>
