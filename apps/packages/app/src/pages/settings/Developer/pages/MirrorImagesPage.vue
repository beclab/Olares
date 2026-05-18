<template>
	<page-title-component :show-back="true" :title="t('Image management')">
		<template v-slot:end>
			<!-- <div
				class="row justify-center items-center"
				:class="deviceStore.isMobile ? '' : 'add-btn'"
				@click="removeImagesConfirm"
			>
				<q-icon
					name="sym_r_delete"
					color="ink-1"
					:size="deviceStore.isMobile ? '32px' : '20px'"
				/>
				<div class="text-body3 add-title" v-if="!deviceStore.isMobile">
					{{ t('Clean up images') }}
				</div>
			</div> -->
		</template>
	</page-title-component>
	<bt-scroll-area class="nav-height-scroll-area-conf">
		<AdaptiveLayout>
			<template v-slot:pc>
				<div class="row items-center full-width">
					<bt-select-v3
						v-model="mirrorName"
						width="100px"
						height="32px"
						:options="mirrorOptions"
					/>
					<q-input
						style="flex: 1"
						class="q-ml-md q-px-md search-input"
						borderless
						dense
						v-model="searchContent"
					>
						<template v-slot:after>
							<q-icon name="sym_r_search" size="16px" color="ink-2" />
						</template>
					</q-input>
				</div>
				<q-list class="q-py-md q-list-class q-mt-md">
					<div
						v-if="filterImages && filterImages.length > 0"
						class="column item-margin-left item-margin-right"
					>
						<q-table
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-2"
							flat
							:bordered="false"
							:rows="filterImages"
							:columns="columns"
							row-key="id"
							hide-pagination
							hide-selected-banner
							hide-bottom
							:rowsPerPageOptions="[0]"
						>
							<template v-slot:body-cell-size="props">
								<q-td
									:props="props"
									class="text-ink-1 text-body1"
									style="height: 64px"
									no-hover
								>
									{{ format.humanStorageSize(props.row.size) }}
								</q-td>
							</template>
							<template v-slot:body-cell-name="props">
								<q-td
									:props="props"
									class="text-ink-1"
									style="height: 64px"
									no-hover
								>
									<div class="text-ink-1 image-name text-body1">
										{{ getImageName(props.row.repo_tags) }}
									</div>
									<div
										class="row q-gutter-xs tags-bg text-body3"
										v-if="props.row.repo_tags && props.row.repo_tags.length > 0"
									>
										<div class="tag-item row items-center justify-center">
											{{ getImageStoreName(props.row.repo_tags[0]) }}
										</div>
									</div>
								</q-td>
							</template>
						</q-table>
					</div>
					<empty-component
						class="q-pb-xl"
						v-else
						:info="t('No image added')"
						:empty-image-top="40"
					/>
				</q-list>
			</template>
			<template v-slot:mobile>
				<div class="row items-center full-width">
					<div style="flex: 1" class="q-mr-sm">
						<bt-select-v3
							v-model="mirrorName"
							height="32px"
							:options="mirrorOptions"
						/>
					</div>
					<q-input
						style="flex: 1"
						class="q-ml-sm q-px-md search-input"
						borderless
						dense
						v-model="searchContent"
					>
						<template v-slot:after>
							<q-icon name="sym_r_search" size="16px" color="ink-2" />
						</template>
					</q-input>
				</div>
				<div v-if="filterImages.length > 0">
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(image, index) in filterImages"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md mobile-image-title-row"
							>
								<div class="mobile-image-title">
									{{ getImageName(image.repo_tags) }}
								</div>
								<!-- <q-icon
									name="sym_r_delete"
									color="ink-2"
									size="20px"
									@click.stop="removeConfirm(image)"
								/> -->
							</div>
						</template>
						<template v-slot:grid>
							<bt-grid-item
								v-if="image.repo_tags && image.repo_tags.length > 0"
								:label="t('Repo name')"
								mobileTitleClasses="text-body3-m"
							>
								<template v-slot:value>
									<div class="text-body3-m mobile-repo-name">
										{{ getImageStoreName(image.repo_tags[0]) }}
									</div>
								</template>
							</bt-grid-item>
							<bt-grid-item
								:label="t('Image size')"
								mobileTitleClasses="text-body3-m"
								:value="format.humanStorageSize(image.size)"
							/>
						</template>
					</bt-grid>
				</div>
				<empty-component
					class="q-pb-xl"
					v-else
					:info="t('No image added')"
					:empty-image-top="40"
				/>
			</template>
		</AdaptiveLayout>
	</bt-scroll-area>
</template>

<script setup lang="ts">
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import BtSelectV3 from 'src/components/settings/base/BtSelectV3.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';
import { useMirrorStore, RegistryImage } from 'src/stores/settings/mirror';
import { computed, onMounted, ref } from 'vue';
import { SelectorProps } from 'src/constant';
import { format } from 'src/utils/format';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
// import BtActionIcon from '../../../../components/settings/base/BtActionIcon.vue';
// import ReminderDialogComponent from '../../../../components/settings/ReminderDialogComponent.vue';
// import { useQuasar } from 'quasar';
// import {
// 	notifyFailed,
// 	notifySuccess
// } from '../../../../utils/settings/btNotify';
// import { useDeviceStore } from '../../../../stores/settings/device';

const { t } = useI18n();
const route = useRoute();
const mirrorStore = useMirrorStore();
const images = ref([] as RegistryImage[]);
const registry = ref((route.query.registry as string) || undefined);
const mirrorName = ref(registry.value ? registry.value : 'all');
const mirrorOptions = ref<SelectorProps[]>([
	{
		label: t('all'),
		value: 'all'
	}
]);
const searchContent = ref('');
// const $q = useQuasar();

// const deviceStore = useDeviceStore();

onMounted(() => {
	getImageList();
});

const getImageList = async () => {
	images.value = await mirrorStore.getRegistryImages();
	const set = new Set<string>();
	images.value.forEach((image) => {
		const storeName = getImageStoreName(image.repo_tags[0]);
		set.add(storeName);
	});
	if (images.value.length > 0) {
		mirrorOptions.value = [
			{
				label: t('all'),
				value: 'all'
			}
		];
		for (const item of set) {
			mirrorOptions.value.push({
				label: item,
				value: item
			});
		}
	}
};

const filterImages = computed(() => {
	return images.value
		.filter((item) => {
			if (mirrorName.value === 'all') {
				return true;
			}
			return getImageStoreName(item.repo_tags[0]) === mirrorName.value;
		})
		.filter((item) => {
			if (searchContent.value) {
				return getImageName(item.repo_tags).includes(searchContent.value);
			} else {
				return true;
			}
		});
});

const columns: any = [
	{
		name: 'name',
		align: 'left',
		label: t('Image name'),
		field: 'repo_tags',
		format: (val: any) => {
			return getImageName(val);
		},
		sortable: false
	},
	{
		name: 'size',
		align: 'right',
		label: t('Image size'),
		field: 'size',
		format: (val: any) => {
			return format.humanStorageSize(val);
		},
		sortable: false
	}
	// {
	// 	name: 'actions',
	// 	align: 'right',
	// 	label: t('action'),
	// 	sortable: false
	// }
];

const getImageName = (repo_tags: string[]) => {
	let itemName = '';
	if (repo_tags && repo_tags.length > 0) {
		itemName =
			registry.value && (repo_tags[0] as string).startsWith(registry.value)
				? (repo_tags[0] as string).substring(registry.value.length + 1)
				: (repo_tags[0] as string);
	}
	if (itemName && itemName.includes('/')) {
		itemName = itemName.substring(itemName.indexOf('/') + 1);
	}
	return itemName;
};

const getImageStoreName = (itemName: string) => {
	if (itemName && itemName.includes('/')) {
		itemName = itemName.split('/')[0];
	}
	return itemName;
};

// const copyImageName = (imageName: string) => {
// copyToClipboard(imageName)
// 	.then(() => {
// 		notifySuccess(t('copy_successfully'));
// 	})
// 	.catch((e) => {
// 		notifyFailed(e.message);
// 	});
// };

// const removeImagesConfirm = () => {
// 	$q.dialog({
// 		component: ReminderDialogComponent,
// 		componentProps: {
// 			title: t('Confirm deletion?'),
// 			message: t('Are you sure you want to clean up unused images?'),
// 			useCancel: true,
// 			confirmText: t('confirm'),
// 			cancelText: t('cancel')
// 		}
// 	}).onOk(async () => {
// 		try {
// 			const data = await mirrorStore.deleteImagesPrune();

// 			if (data && data.count > 0) {
// 				notifySuccess(
// 					t('Successfully deleted {images} images, freeing up {size} space', {
// 						images: data?.count || 0,
// 						size: format.humanStorageSize(data?.size || 0)
// 					})
// 				);
// 			} else {
// 				notifySuccess(t('successful'));
// 			}
// 		} catch (error) {
// 			notifyFailed(error);
// 		}
// 	});
// };
</script>

<style scoped lang="scss">
.search-input {
	border: 1px solid $input-stroke;
	border-radius: 8px;
}

.add-btn {
	border-radius: 8px;
	padding: 6px 12px;
	border: 1px solid $separator;
	cursor: pointer;
	text-decoration: none;

	.add-title {
		color: $ink-2;
	}
}

.add-btn:hover {
	background-color: $background-3;
}

.tags-bg {
	width: 100%;
	// background-color: red;
	// margin-top: 4px;
	height: auto;

	.tag-item {
		// padding: 0px 8px;
		// border-radius: 4px;
		// height: 25px;
		// margin: 0;
		// margin-top: 0px;
		// background-color: $background-3;

		color: $ink-3;

		.icon {
			width: 12px;
			height: 12px;
			margin-left: 8px;
			color: $ink-2;
		}

		// &:hover {
		// 	background-color: $background-5;
		// }

		/* 120% */
	}
}

.image-name {
	max-width: 460px;
	text-overflow: ellipsis;
	white-space: nowrap;
	overflow: hidden;
}

.mobile-image-title-row {
	min-width: 0;
}

.mobile-image-title,
.mobile-repo-name {
	min-width: 0;
	white-space: normal;
	overflow-wrap: anywhere;
	word-break: break-word;
}

::v-deep(
		.q-field--dense .q-field__control,
		.q-field--dense .q-field__marginal
	) {
	height: 30px;
}

::v-deep(.q-field--dense .q-field__after, .q-field--dense .q-field__append) {
	height: 30px;
}
</style>
