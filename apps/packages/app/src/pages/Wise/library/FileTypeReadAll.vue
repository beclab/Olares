<template>
	<q-btn
		class="btn-size-sm btn-no-text btn-no-border"
		color="ink-2"
		outline
		no-caps
		:loading="loading"
		:icon="!readAll ? 'sym_r_format_list_bulleted' : 'sym_r_checklist_rtl'"
		@click="setReadAll()"
	>
		<template v-slot:loading>
			<bt-loading :loading="loading" />
		</template>
		<bt-tooltip
			:label="!readAll ? t('main.mask_all_unseen') : t('main.mask_all_seen')"
		/>
	</q-btn>
</template>
<script setup lang="ts">
import BtTooltip from '../../../components/base/BtTooltip.vue';
import { FILE_TYPE } from '../../../utils/rss-types';
import BtLoading from '../../../components/base/BtLoading.vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { useRssStore } from '../../../stores/rss';
import { PropType, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { sendMessageToWorker } from '../database/sqliteService';

const props = defineProps({
	fileType: {
		type: Object as PropType<FILE_TYPE>,
		required: true
	},
	readAll: {
		type: Boolean,
		default: true
	}
});

const { t } = useI18n();
const rssStore = useRssStore();
const loading = ref(false);

const setReadAll = async () => {
	loading.value = true;
	sendMessageToWorker('query', {
		sql: `SELECT entries.* FROM entries CROSS JOIN json_each(sources) WHERE json_each.value = 'wise' AND unread = ${props.readAll};`
	}).then((data: any) => {
		const ids = data.map((item) => {
			return item.id;
		});
		if (ids.length > 0) {
			console.log(ids);
			rssStore.markEntryUnread(ids, !props.readAll).finally(() => {
				loading.value = false;
			});
		} else {
			BtNotify.show({
				type: NotifyDefinedType.FAILED,
				message: t('base.no_matching_content')
			});
		}
	});
};
</script>

<style scoped lang="scss"></style>
