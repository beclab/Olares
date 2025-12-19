<template>
	<BaseSiteCard :data="info">
		<template #action>
			<q-btn
				class="open-file-wrapper"
				padding="6px"
				v-if="data.exist"
				@click="openWise(data.exist_entry_id)"
			>
				<q-icon
					name="sym_r_open_in_new"
					:color="theme?.btnTextActiveColor"
					size="20px"
				/>
			</q-btn>
			<q-btn
				:color="theme?.btnDefaultColor"
				padding="6px"
				:loading="collectSiteStore.collectLoading"
				@click="clickHandler(data.url)"
				v-else
				:disable="data.disabled"
			>
				<q-icon
					name="sym_r_box_add"
					:color="theme?.btnTextDefaultColor"
					size="20px"
				/>
			</q-btn>
		</template>
		<TagEditPopup2 :entry="data" :disabled="collectSiteStore.loading" />
	</BaseSiteCard>
</template>

<script setup lang="ts">
import { computed, inject, ref } from 'vue';
import BaseSiteCard from '../../components/collection/BaseSiteCard.vue';
import { CollectEntry } from 'src/types/commonApi';
import { useCollectSiteStore } from 'src/stores/collect-site';
import TagEditPopup2 from 'src/components/rss/TagEditPopup2.vue';
import { useUserStore } from 'src/stores/user';
import { openUrl } from 'src/utils/bex/tabs';
import { useAppAbilitiesStore } from 'src/stores/appAbilities';
import { useRouter } from 'vue-router';
import { COLLECT_THEME } from 'src/constant/provide';
import { COLLECT_THEME_TYPE } from 'src/constant/theme';
const props = defineProps<{ data: CollectEntry & { disabled?: boolean } }>();

const theme = inject<COLLECT_THEME_TYPE>(COLLECT_THEME);

const collectSiteStore = useCollectSiteStore();
const userStore = useUserStore();

const clickHandler = (url: string) => {
	collectSiteStore.addCollect(url);
};
const appAbilitiesStore = useAppAbilitiesStore();
const router = useRouter();

const openWise = (id?: string) => {
	if (!id) {
		return;
	}

	const wiseEntryPath = `/history/${id}`;
	if (process.env.APPLICATION !== 'WISE') {
		const url = userStore.getModuleSever(
			appAbilitiesStore.wise.id,
			'https:',
			wiseEntryPath
		);
		if (!url) {
			return;
		}
		openUrl(url);
	} else {
		openUrl(`${location.origin}${wiseEntryPath}`);
	}
};

const info = computed(() => {
	return {
		...props.data,
		icon: props.data.thumbnail
	};
});
</script>

<style lang="scss" scoped>
.open-file-wrapper {
	border: 1px solid $btn-stroke;
}
</style>
