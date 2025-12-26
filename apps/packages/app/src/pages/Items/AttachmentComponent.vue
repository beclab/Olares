<template>
	<div class="column attach">
		<div class="">
			<div class="text-light-blue-default text-body3">
				{{ attach?.name }}
			</div>
			<div class="text-body3 text-ink-1">
				{{
					(attach.type || t('vault_t.unkown_file_type')) +
					' - ' +
					format.formatFileSize(attach?.size)
				}}
			</div>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { ref, watch } from 'vue';
import { AttachmentInfo } from '@didvault/sdk/src/core';
// import { fileSize } from '@didvault/sdk/src/util';
import { format } from '../../utils/format';
import { useI18n } from 'vue-i18n';

const props = defineProps({
	itemID: {
		type: String,
		required: true
	},
	attach: {
		type: AttachmentInfo,
		required: true
	},
	editing: {
		type: Boolean,
		required: true
	}
});

const attachVaule = ref(props.attach.name);

watch(
	() => props.attach,
	(newValue, oldValue) => {
		if (newValue == oldValue) {
			return;
		}
		attachVaule.value = props.attach.name;
	}
);

const { t } = useI18n();
</script>

<style lang="scss" scoped>
.attach {
	padding-left: 5px;
}
</style>
