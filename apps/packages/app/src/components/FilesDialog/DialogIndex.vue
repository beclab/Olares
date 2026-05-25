<template>
	<component
		:origin_id="origin_id"
		:is="currentView"
		:selectType="selectType"
		:origins="origins"
		@onSubmit="onSubmit"
	/>
</template>

<script setup lang="ts">
import { onMounted, ref, PropType } from 'vue';
import { useQuasar } from 'quasar';
import { isPad } from '../../utils/platform';
import { PickType, useFilesStore } from '../../stores/files';
import { DriveType } from '../../utils/interface/files';
import DialogForder from './DialogForder.vue';
import DialogFolderMobile from './DialogFolderMobile.vue';

const props = defineProps({
	selectType: {
		type: String as PropType<PickType>,
		required: false,
		default: PickType.FOLDER
	},
	masterNode: {
		type: Boolean,
		required: false,
		default: false
	},

	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			return [
				DriveType.Drive,
				DriveType.Sync,
				DriveType.External,
				DriveType.Cache,
				DriveType.Data,
				DriveType.GoogleDrive
			];
		}
	}
});

console.log('propsorigins', props.origins);

const emits = defineEmits(['onSubmit']);

const filesStore = useFilesStore();
const $q = useQuasar();
const currentView = ref();
const origin_id = ref(Date.now());

filesStore.initIdState(origin_id.value);
filesStore.onlyMasterNodes[origin_id.value] = props.masterNode;

const onSubmit = (value) => {
	emits('onSubmit', value);
};

const isMobile = ref(
	(process.env.PLATFORM == 'MOBILE' || $q.platform.is.mobile) && !isPad()
);

onMounted(() => {
	if (isMobile.value) {
		currentView.value = DialogFolderMobile;
	} else {
		currentView.value = DialogForder;
	}
});
</script>

<style lang="scss" scoped></style>
