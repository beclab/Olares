<template>
	<span>
		<template v-if="reason">
			<StatusReason :status="status" :reason="reason" :data="detail" />
		</template>
		<template v-else-if="detail.updateTime">
			{{
				t('UPDATED_TIME', {
					value: getLocalTime(detail.updateTime).format('YYYY-MM-DD HH:mm:ss')
				})
			}}
		</template>
		<template v-else>
			{{
				t('CREATED_TIME', {
					diff: getLocalTime(detail.createTime).format('YYYY-MM-DD HH:mm:ss')
				})
			}}
		</template>
	</span>
</template>

<script setup lang="ts">
import { defineProps } from 'vue';
import { getLocalTime } from '@apps/control-panel-common/src/utils';
import { getWorkloadStatus } from '@apps/control-panel-common/src/utils/status';
import StatusReason from '../components/StatusReason.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();
interface DescriptionProps {
	detail: any;
}

const props = defineProps<DescriptionProps>();

const { status, reason } = getWorkloadStatus(props.detail, props.detail.module);
</script>
