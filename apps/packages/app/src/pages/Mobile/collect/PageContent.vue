<template>
	<div class="page-content">
		<EmptyData
			v-if="!item.url"
			:title="$t('no_data')"
			btn-hidden
			size="sm"
		></EmptyData>
		<collect-item
			:item="item"
			style="padding: 0"
			:outline="false"
			:link="false"
			contentClass="item-content-wrapper"
			v-else
		>
			<template v-slot:image>
				<div class="image-avatar-container">
					<q-img
						:src="
							item.image
								? item.image
								: getRequireImage('rss/page_default_img.svg')
						"
						class="image-avatar"
					>
						<template v-slot:loading>
							<q-img
								:src="getRequireImage('rss/page_default_img.svg')"
								class="image-avatar"
							/>
						</template>
						<template v-slot:error>
							<q-img
								:src="getRequireImage('rss/page_default_img.svg')"
								class="image-avatar"
							/>
						</template>
					</q-img>
				</div>
			</template>
			<template v-slot:side>
				<CustomButton
					v-if="item.status === RssStatus.added"
					color="yellow-default"
					class="full-width"
					@click="openWise"
					:disable="!appAbilitiesStore.wise.running"
				>
					<template #label>
						<div class="text-ink-on-brand-black row items-center">
							<q-icon name="sym_r_open_in_new" size="20px" />
							<span class="q-ml-sm">{{ $t('bex.open_in_wise') }} </span>
						</div>
					</template>
				</CustomButton>
				<CustomButton
					v-else
					color="yellow-default"
					class="full-width"
					@click="onSaveEntry(item)"
					:disable="!appAbilitiesStore.wise.running"
				>
					<template #label>
						<div class="row items-center">
							<q-icon name="sym_r_box_add" size="20px" />
							<span class="q-ml-sm">{{ $t('Collect to wise') }}</span>
						</div>
					</template>
				</CustomButton>
			</template>
		</collect-item>
	</div>
</template>

<script setup lang="ts">
import CollectionItemStatus from './CollectionItemStatus.vue';
import CollectItem from './CollectItem.vue';
import { getRequireImage } from '../../../utils/imageUtils';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';
import { useCollect } from 'src/composables/bex/useCollect';
import { onMounted, onUnmounted } from 'vue';
import EmptyData from 'src/pages/Plugin/components/EmptyData.vue';
import { browser } from 'webextension-polyfill-ts';

const {
	collectStore,
	RssStatus,
	onSaveEntry,
	openWise,
	item,
	setData,
	init,
	appAbilitiesStore,
	handleActivated,
	handleUpdated
} = useCollect();

onMounted(() => {
	init();
	browser.tabs.onActivated.addListener(handleActivated);
	browser.tabs.onUpdated.addListener(handleUpdated);
});

onUnmounted(() => {
	browser.tabs.onActivated.removeListener(handleActivated);
	browser.tabs.onUpdated.removeListener(handleUpdated);
});
</script>

<style scoped lang="scss">
.page-content {
	width: 100%;
	height: 100%;
	.image-avatar-container {
		width: 60px;
		height: 60px;
		padding: 8px;
		border-radius: 12px;
		border: 1px solid red;
		border: 0.938px solid $separator-2;
		overflow: hidden;
		background: $background-1;

		.image-avatar {
			width: 100%;
		}
	}
	::v-deep(.item-content-wrapper) {
		padding: 0;
	}
}
</style>
