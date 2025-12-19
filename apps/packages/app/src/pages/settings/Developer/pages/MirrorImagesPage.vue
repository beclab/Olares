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
				<q-list class="q-py-md q-list-class q-mt-md">
					<div
						v-if="images && images.length > 0"
						class="column item-margin-left item-margin-right"
					>
						<q-table
							tableHeaderStyle="height: 32px;"
							table-header-class="text-body3 text-ink-2"
							flat
							:bordered="false"
							:rows="images"
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
				<div v-if="images.length > 0">
					<bt-grid
						class="mobile-items-list"
						:repeat-count="2"
						v-for="(image, index) in images"
						:key="index"
						:paddingY="12"
					>
						<template v-slot:title>
							<div
								class="text-subtitle3-m row justify-between items-center clickable-view q-mb-md"
							>
								<div>
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
								:value="getImageStoreName(image.repo_tags[0])"
							>
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
import { onMounted, ref } from 'vue';
import PageTitleComponent from 'src/components/settings/PageTitleComponent.vue';
import { useI18n } from 'vue-i18n';
import AdaptiveLayout from 'src/components/settings/AdaptiveLayout.vue';
import BtGridItem from 'src/components/settings/base/BtGridItem.vue';
import BtGrid from 'src/components/settings/base/BtGrid.vue';

import EmptyComponent from 'src/components/settings/EmptyComponent.vue';
import {
	useMirrorStore,
	RegistryImage
} from '../../../../stores/settings/mirror';
// import BtActionIcon from '../../../../components/settings/base/BtActionIcon.vue';

import { useRoute } from 'vue-router';
// import ReminderDialogComponent from '../../../../components/settings/ReminderDialogComponent.vue';
// import { useQuasar } from 'quasar';
import { format } from '../../../../utils/format';
// import {
// 	notifyFailed,
// 	notifySuccess
// } from '../../../../utils/settings/btNotify';
// import { useDeviceStore } from '../../../../stores/settings/device';

const { t } = useI18n();

const mirrorStore = useMirrorStore();

const route = useRoute();

const registry = ref((route.query.registry as string) || undefined);

const images = ref([] as RegistryImage[]);

// const $q = useQuasar();

// const deviceStore = useDeviceStore();

onMounted(async () => {
	getImageList();
});

const getImageList = async () => {
	images.value = await mirrorStore.getRegistryImages(registry.value);
};

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
</style>
