<template>
	<MyGridLayout :col-width="colWidth" gap="xl">
		<div
			class="col-6 col-md-4 col-lg-3 my-list-content row items-center no-wrap"
			style="white-space: nowrap"
			v-for="item in data"
			:key="item.name"
		>
			<div class="text-body3 text-ink-3">{{ item.name }}：</div>
			<div class="text-subtitle2 text-ink-1">
				<div v-if="item.name === 'Cluster'">
					<span>{{ Cluster }}</span>
				</div>
				<div v-else-if="item.name === t('STATUS')" class="row items-center">
					<MyBadge :type="item.type || item.value" class="q-mr-xs"></MyBadge>
					<span>{{ item.value }}</span>
				</div>
				<div v-else-if="item.name === 'Endpoint'">
					{{ EndpointList(item.value) }}
				</div>
				<div v-else-if="item.name === $t('PASSWORD')" class="row items-end">
					<slot v-if="$slots.Password" name="Password" :data="item"></slot>
					<div v-else>
						<PasswordToggle :value="item.value"></PasswordToggle>
					</div>
				</div>
				<div v-else class="row items-center flex-gap-x-xs">
					{{ isNil(item.value) || item.value === '' ? '-' : item.value }}
					<q-icon v-if="item.icon" v-bind="item.icon" size="16px" />
				</div>
			</div>
		</div>
	</MyGridLayout>
</template>

<script setup lang="ts">
import { isArray, isEmpty, isNil, snakeCase } from 'lodash';
import Status from '../components/Status.vue';
import passwordEditIcon from '../assets/password-edit.svg';
import MyGridLayout from '../components/MyGridLayout.vue';
import MyBadge from '../components/MyBadge.vue';
import PasswordToggle from '../components/PasswordToggle.vue';
import { useI18n } from 'vue-i18n';
const { t } = useI18n();

interface Props {
	data:
		| Array<{
				name: string;
				value: string | number;
				type?: string;
				icon?: {
					name: string;
					color?: string;
				};
		  }>
		| undefined;
	colWidth?: string;
}
const props = withDefaults(defineProps<Props>(), {});

const EndpointList = (value: any) => {
	return isArray(value) ? value.join(' ') : value;
};

const Cluster = 'default';
</script>

<style lang="scss" scoped>
.my-list-content {
	word-break: break-all;
	white-space: pre-wrap;
	font-weight: 500;
	line-height: 16px;
}
</style>
