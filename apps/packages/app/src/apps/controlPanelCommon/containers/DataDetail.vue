<template>
	<div class="my-table-container" v-if="hasData">
		<table class="my-data-detail-table">
			<tr v-for="(value, key) in data" :key="key + value" class="table-wrapper">
				<td class="first-td q-pl-lg q-pr-lg q-py-md">
					<div>
						{{ key }}
					</div>
				</td>
				<td class="q-pr-lg q-pl-md q-py-md">
					<MyEllips :text="value"></MyEllips>
				</td>
			</tr>
		</table>
	</div>
	<Empty v-else-if="!noEmpty" size="small"></Empty>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import MyEllips from '../components/MyEllips.vue';
import Empty from '../components/Empty.vue';
import { isEmpty } from 'lodash';

interface Props {
	data: { [key: string]: any };
	noEmpty: boolean;
}

const props = withDefaults(defineProps<Props>(), {});
const hasData = computed(() => {
	return !isEmpty(props.data);
});
</script>

<style lang="scss" scoped>
.my-table-container {
	overflow-y: hidden;
}
.my-data-detail-table {
	overflow-wrap: anywhere;
	border-collapse: separate;
	border-spacing: 0 12px;
	margin: -12px 0;
	width: 100%;
	.first-td {
		vertical-align: baseline;
		max-width: 30%;
	}
	& tr td {
		border: 1px solid $separator;
	}
	& tr.table-wrapper td:first-child {
		border-radius: 12px 0 0 12px;
		white-space: nowrap;
		border-right: none;
	}
	& tr.table-wrapper td:last-child {
		border-radius: 0 12px 12px 0;
		border-left: none;
	}
	.text-wrap {
		white-space: wrap;
	}
}
</style>
