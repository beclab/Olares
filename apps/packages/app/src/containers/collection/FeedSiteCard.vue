<template>
	<BaseSiteCard :data="info">
		<template #action>
			<q-btn
				:color="
					feed.is_subscribed
						? theme?.btnFeedDefaultColor
						: theme?.btnDefaultColor
				"
				padding="6px"
				:disable="feed.is_subscribed || feed.disabled"
				:loading="feed.loading"
				@click="onOK"
			>
				<q-icon
					v-if="feed.is_subscribed"
					name="sym_r_bookmark_added"
					:color="theme?.btnTextFeedActiveColor"
				/>
				<q-icon
					v-else
					name="sym_r_bookmark_add"
					:color="theme?.btnTextDefaultColor"
					size="20px"
				/>
			</q-btn>
		</template>
	</BaseSiteCard>
</template>

<script setup lang="ts">
import BaseSiteCard from '../../components/collection/BaseSiteCard.vue';
import { computed, inject, ref, toRefs } from 'vue';
import { FeedItem } from 'src/types/commonApi';
import { useCollectSiteStore } from 'src/stores/collect-site';
import { handleSiteIcon } from 'src/utils/image';
import { COLLECT_THEME } from 'src/constant/provide';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';

interface Props {
	feed: FeedItem & { disabled?: boolean };
}
const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);

const collectSiteStore = useCollectSiteStore();
const props = withDefaults(defineProps<Props>(), {});

const info = computed(() => {
	return {
		id: props.feed.id,
		title: props.feed.title,
		url: props.feed.feed_url,
		icon: handleSiteIcon(props.feed.icon_content, props.feed.icon_type)
	};
});
const onOK = () => {
	collectSiteStore.addFeed(props.feed.feed_url);
};
</script>

<style></style>
