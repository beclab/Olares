<template>
	<div class="pdf-content">
		<collect-item
			v-for="(item, index) in collectStore.pdfList"
			:key="index"
			:item="item"
			class="q-mt-md"
		>
			<template v-slot:image>
				<q-img
					:src="getRequireImage('rss/rss_pdf_img.svg')"
					class="image-avatar"
				/>
			</template>
			<template v-slot:side>
				<CollectionItemStatus class="status-white-bg">
					<template v-slot:status>
						<q-icon
							v-if="item.status === PDFStatus.none"
							name="sym_r_add_box"
							size="20px"
							class="text-ink-1 cursor-pointer"
							@click="onDownloadPdf(item)"
						/>
						<q-knob
							v-if="item.status === PDFStatus.loading"
							:model-value="
								item.progress ? item.progress.download / item.progress.total : 0
							"
							size="20px"
							:min="0"
							:max="1"
							:thickness="0.22"
							color="yellow-7"
							track-color="grey-1"
							class=""
						/>
						<q-icon
							v-if="item.status === PDFStatus.error"
							name="sym_r_error"
							size="20px"
							class="text-negative"
						/>
						<q-icon
							v-if="item.status === PDFStatus.success"
							name="sym_r_check_circle"
							size="20px"
							class="text-positive"
						/>
					</template>
				</CollectionItemStatus>
			</template>
		</collect-item>
	</div>
</template>

<script setup lang="ts">
import CollectItem from './CollectItem.vue';
import { PDFInfo, PDFStatus } from './utils';
import CollectionItemStatus from './CollectionItemStatus.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import { BtDialog } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import { useCollectStore } from '../../../stores/collect';

const { t } = useI18n();
const collectStore = useCollectStore();

const onDownloadPdf = (item: PDFInfo) => {
	const array = item.url.split('/');
	let fileName = '';
	if (array.length > 0) {
		fileName = array[array.length - 1];
	}
	BtDialog.show({
		title: t('rename_file'),
		okStyle: {
			background: '#FF7733',
			color: '#ffffff'
		},
		cancel: true,
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		prompt: {
			model: fileName,
			type: 'text', // optional
			name: t('pdf_name'),
			placeholder: t('input_pdf_name')
		}
	})
		.then((name: any) => {
			if (name) {
				collectStore.downloadPdf(item.url, name).then((result) => {
					if (result) {
						item.id = result;
						item.status = PDFStatus.loading;
					}
				});
			}
		})
		.catch((err: Error) => {
			console.log('click ok', err);
		});
};
</script>

<style scoped lang="scss">
.pdf-content {
	width: 100%;
	height: 100%;

	.image-avatar {
		width: 44px;
		height: 44px;
		border-radius: 8px;
	}

	.status-white-bg {
		background: #fff;
	}
}
</style>
