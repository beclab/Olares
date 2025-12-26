<template>
	<div class="files-shard-root">
		<terminus-title-bar
			:title="title"
			:is-dark="isDark"
			:hookBackAction="true"
			@on-return-action="back"
		>
			<template v-slot:right>
				<div class="row items-center" v-if="origin_id === FilesIdType.PAGEID">
					<q-btn
						class="text-ink-1 btn-size-sm btn-no-text btn-no-border"
						icon="sym_r_close"
						text-color="ink-2"
						@click="closeDialog"
					>
					</q-btn>
				</div>
			</template>
		</terminus-title-bar>
		<div class="content q-px-lg">
			<template v-if="recentFolder.length > 0">
				<div class="text-subtitle2 text-ink-3">{{ t('recent_folder') }}</div>
				<div
					class="shard-item row items-center justify-between"
					v-for="(item, index) in recentFolder"
					:key="item.name"
					@click="handleSelect(index)"
				>
					<div class="row items-center justify-center">
						<img src="/img/folder-default.svg" alt="icon" />
						<div>
							<div class="recent-name text-subtitle2 text-ink-1 q-ml-md">
								{{ item.name }}
							</div>
							<div class="recent-path text-body3 text-ink-3 q-ml-md">
								{{ item.path }}
							</div>
						</div>
					</div>
					<div
						class="column justify-center items-center"
						style="height: 32px; width: 32px"
						v-if="selectedIndex === index"
					>
						<img src="/img/files-selected.svg" alt="icon" />
					</div>
				</div>
			</template>

			<div class="text-subtitle2 text-ink-3">
				{{ t('upload_to_location') }}
			</div>
			<div
				class="shard-item row items-center justify-between"
				v-for="menu in sharedMenus"
				:key="menu"
				@click="open(menu)"
			>
				<div class="row items-center justify-center">
					<img :src="getShardIcon(menu.icon)" alt="icon" />
					<div class="text-subtitle2 text-ink-1 q-ml-md">{{ menu.label }}</div>
				</div>
				<div
					class="column justify-center items-center"
					style="height: 32px; width: 32px"
				>
					<q-icon name="sym_r_keyboard_arrow_right" size="20px" class="grey-8">
					</q-icon>
				</div>
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
// interface LocationType {
// 	name: string;
// 	icon: string;
// }

import { ref, onMounted, PropType } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { MenuItem } from '../../../utils/contact';
import {
	useFilesStore,
	FilesIdType,
	MenuItemType
} from './../../../stores/files';
import TerminusTitleBar from '../../../components/common/TerminusTitleBar.vue';
import { DriveType } from '../../../utils/interface/files';
import { filesIsV2 } from 'src/api';
import { FileSharedService } from 'src/platform/interface/capacitor/plugins/share';

const props = defineProps({
	origins: {
		type: Array as PropType<DriveType[]>,
		required: false,
		default: () => {
			if (filesIsV2()) {
				return [DriveType.Drive, DriveType.External];
			}
			return [DriveType.Drive, DriveType.External, DriveType.Sync];
		}
	},
	origin_id: {
		type: Number,
		required: false
	}
});

const emits = defineEmits(['open', 'close']);

const recentFolder = ref<any[]>([]);
const filesStore = useFilesStore();
const selectedIndex = ref<number>(0);
const router = useRouter();

const isDark = ref(false);
const { t } = useI18n();

const title = ref(t('select_folder_to_upload'));
const sharedMenus = ref();

const closeDialog = () => {
	emits('close');
};

const getShardIcon = (img: string) => {
	return `/img/${img}`;
};

const handleSelect = (index: number) => {
	selectedIndex.value = index;
};

const open = (item: MenuItemType) => {
	if (props.origin_id) {
		emits('open', item);
		return;
	}

	if (item.driveType === DriveType.Drive) {
		const url = `/Files/Home/`;
		filesStore.setBrowserUrl(url, DriveType.Drive);
	} else if (item.driveType == DriveType.External) {
		filesStore.setBrowserUrl('/Files/External/', DriveType.External);
	} else if (item.driveType === DriveType.Sync) {
		router.push({
			path: `/repo/${MenuItem.MYLIBRARIES}/`
		});
	}
};

onMounted(async () => {
	sharedMenus.value = await filesStore.getMobileMenu(props.origins);

	const data: any[] = localStorage.getItem('recentFolder')
		? JSON.parse(localStorage.getItem('recentFolder') as string)
		: [];
	recentFolder.value = data;
});

const back = () => {
	if (props.origin_id) {
		emits('close');
		return false;
	}
	filesStore.isShard = false;
	FileSharedService.reset();
	router.back();
};
</script>

<style lang="scss" scoped>
.files-shard-root {
	width: 100%;
	height: 100%;

	.content {
		width: 100%;
		height: calc(100% - 56px);

		.shard-item {
			height: 72px;
			line-height: 72px;
			border-bottom: 1px solid rgba($color: #000000, $alpha: 0.1);

			.recent-name,
			.recent-path {
				max-width: 200px;
				text-overflow: ellipsis;
				white-space: nowrap;
				overflow: hidden;
			}
		}
	}
}
</style>
