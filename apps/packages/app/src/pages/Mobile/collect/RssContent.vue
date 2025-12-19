<template>
	<div class="rss-content column flex-gap-y-lg">
		<collect-item
			v-for="(item, index) in collectStore.rssList"
			:key="index"
			:item="item"
			outline
			link
		>
			<template v-slot:image>
				<div>
					<FeedIcon :feed="item.feed" size="40px"></FeedIcon>
				</div>
			</template>
			<template v-slot:side>
				<CustomButton
					v-if="item.status === RssStatus.added"
					color="background-3"
					class="full-width"
					@click="onRemoveFeed(item)"
				>
					<template #label>
						<div class="text-ink-3 row items-center">
							<q-icon name="sym_r_bookmark_add" size="20px" />
							<span class="q-ml-sm">{{ $t('bex.subscribed') }}</span>
						</div>
					</template>
				</CustomButton>

				<CustomButton
					v-else
					color="yellow-default"
					class="full-width"
					@click="onSaveFeed(item)"
				>
					<template #label>
						<div class="row items-center">
							<q-icon name="sym_r_bookmark_added" size="20px" />
							<span class="q-ml-sm">{{ $t('bex.subscribe') }}</span>
						</div>
					</template>
				</CustomButton>
			</template>
		</collect-item>
	</div>
</template>

<script setup lang="ts">
import { RssInfo, RssStatus } from './utils';
import CollectionItemStatus from './CollectionItemStatus.vue';
import CollectItem from './CollectItem.vue';
import { useCollectStore } from '../../../stores/collect';
import { useQuasar } from 'quasar';
import { BtDialog, useColor } from '@bytetrade/ui';
import { useI18n } from 'vue-i18n';
import FeedIcon from '../../../components/rss/FeedIcon.vue';
import CustomButton from 'src/pages/Plugin/components/CustomButton.vue';

const { t } = useI18n();
const { color: orange } = useColor('yellow-default');
const { color: textInk } = useColor('ink-2');

const collectStore = useCollectStore();
const $q = useQuasar();

const onSaveFeed = async (item: RssInfo) => {
	$q.loading.show();
	await collectStore.addFeed(item);
	$q.loading.hide();
};

const onRemoveFeed = async (item: RssInfo) => {
	BtDialog.show({
		title: t('dialog.remove_subscription'),
		message: t('dialog.remove_subscription_desc'),
		okStyle: {
			background: orange.value,
			color: textInk.value
		},
		okText: t('base.confirm'),
		cancelText: t('base.cancel'),
		cancel: true
	})
		.then(async (res) => {
			if (res) {
				$q.loading.show();
				await collectStore.deleteRss([item.url]);
				$q.loading.hide();
			} else {
				console.log('click cancel');
			}
		})
		.catch((err) => {
			console.log(err);
			$q.loading.hide();
		});
};
</script>

<style scoped lang="scss">
.rss-content {
}
</style>
