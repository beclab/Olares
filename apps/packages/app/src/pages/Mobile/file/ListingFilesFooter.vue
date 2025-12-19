<template>
	<div class="footer full-width row items-center justify-around">
		<div @click="createFolder">{{ t('files_popup_menu.new_folder') }}</div>
		<div :class="btnClass" @click="handleUpload">
			{{ selectType ? t('confirm') : t('upload_here') }}
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType, computed } from 'vue';
import { useQuasar } from 'quasar';
import { useFilesStore, FilesIdType, PickType } from './../../../stores/files';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import NewDir from '.././../../components/files/prompts/NewDir.vue';
import { formatFilePath } from '../../../constant';
import { FileSharedService } from 'src/platform/interface/capacitor/plugins/share';

const emits = defineEmits(['onSubmit']);

const props = defineProps({
	origin_id: {
		type: Number,
		required: false,
		default: FilesIdType.PAGEID
	},
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: ''
	}
});

const $q = useQuasar();
const { t } = useI18n();
const router = useRouter();
const filesStore = useFilesStore();

const btnClass = computed(() => {
	if (process.env.APPLICATION === 'WISE') {
		return 'wise-global-ok-button';
	} else if (process.env.APPLICATION === 'SETTINGS') {
		return 'settings-global-ok-button';
	} else {
		return '';
	}
});

const createFolder = () => {
	$q.dialog({
		component: NewDir,
		componentProps: {
			origin_id: props.origin_id
		}
	});
};

const handleUpload = async () => {
	if (props.origin_id) {
		const formatData = formatFilePath(filesStore.currentPath[props.origin_id]);
		emits('onSubmit', formatData);
		return false;
	}

	FileSharedService.sharedFiles({
		target: {
			driveType: filesStore.currentPath[props.origin_id].driveType,
			path: filesStore.currentPath[props.origin_id].path,
			params: filesStore.currentPath[props.origin_id].param
		}
	});
	filesStore.backStack[props.origin_id] = [];
	filesStore.isShard = false;
	router.replace('/home');
};
</script>

<style scoped lang="scss">
.footer {
	position: fixed;
	bottom: 0;
	left: 0;
	height: 88px;
	background-color: #ffffff;

	div {
		width: 45%;
		height: 48px;
		line-height: 48px;
		text-align: center;
		border-radius: 8px;
		box-sizing: border-box;

		&:first-child {
			border: 1px solid rgba(0, 0, 0, 0.1);
		}

		&:last-child {
			background-color: $yellow-default;
		}
	}
}
</style>
