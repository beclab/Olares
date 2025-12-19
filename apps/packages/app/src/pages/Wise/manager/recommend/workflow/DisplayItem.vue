<template>
	<div class="display-item-root row justify-start items-center">
		<div class="display-item-title text-body2 text-ink-3">
			{{ title }}
		</div>
		<div
			v-if="!env"
			class="text-body2 text-ink-2"
			:style="{ width: copy ? 'calc(60% - 25px)' : '60%' }"
		>
			{{ content }}
		</div>
		<q-icon
			class="cursor-pointer"
			v-if="copy && !env"
			size="20px"
			color="ink-2"
			style="margin-left: 5px"
			name="sym_r_file_copy"
			@click="onCopy"
		/>
		<div class="column justify-start text-body2 text-ink-2" style="width: 60%">
			<template v-for="(item, index) in env" :key="'index' + index">
				<div v-if="item.value" style="margin-top: 10px">
					{{ item.name + ' = ' + item.value }}
				</div>
				<div v-if="item.valueFrom" style="margin-top: 10px">
					{{ item.name }}
				</div>
				<div v-if="item.valueFrom" style="margin-left: 20px">
					{{
						item.valueFrom.configMapKeyRef.name +
						' = ' +
						item.valueFrom.configMapKeyRef.key
					}}
				</div>
			</template>
		</div>
	</div>
</template>

<script lang="ts" setup>
import { PropType } from 'vue';
import { Env } from 'src/utils/rss-types';
import { useI18n } from 'vue-i18n';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import { getApplication } from '../../../../../application/base';

const { t } = useI18n();

const props = defineProps({
	env: {
		type: Object as PropType<Env[]>,
		require: false
	},
	title: {
		type: String,
		required: true
	},
	content: {
		type: String,
		required: false
	},
	copy: {
		type: Boolean,
		default: false
	}
});

const onCopy = () => {
	if (props.content) {
		getApplication()
			.copyToClipboard(props.content)
			.then(() => {
				BtNotify.show({
					type: NotifyDefinedType.SUCCESS,
					message: t('copy_success')
				});
			})
			.catch((e) => {
				BtNotify.show({
					type: NotifyDefinedType.FAILED,
					message: t('copy_failure_message', e.message)
				});
			});
	}
};
</script>

<style lang="scss" scoped>
.display-item-root {
	width: 100%;
	height: auto;
	min-height: 40px;
	overflow: hidden;
	margin-top: 20px;

	.display-item-title {
		width: 40%;
	}
}
</style>
