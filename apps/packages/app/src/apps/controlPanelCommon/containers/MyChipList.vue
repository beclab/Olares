<template>
	<div class="my-chip-list-container">
		<table class="my-data-detail-table">
			<tr v-for="item in data" :key="item.name" class="table-wrapper">
				<td class="first-td q-py-md">
					<div>
						{{ item.name }}
					</div>
				</td>
				<td class="q-py-md">
					<MyEllips :text="valueFormat(item)"></MyEllips>
				</td>
			</tr>
		</table>
	</div>
</template>

<script setup lang="ts">
import { lowerCase } from 'lodash';
import MyEllips from '../components/MyEllips.vue';
interface Props {
	data: { [key: string]: any };
}

withDefaults(defineProps<Props>(), {});

const valueFormat = (item: { name: string; value: string }) => {
	return lowerCase(item.name).includes('secret') ||
		lowerCase(item.name).includes('secret')
		? '******'
		: item.value;
};
</script>

<style lang="scss" scoped>
.my-chip-list-container {
	overflow-y: hidden;
}
.my-data-detail-table {
	overflow-wrap: anywhere;
	border-collapse: separate;
	border-spacing: 0;
	width: 100%;
	.first-td {
		vertical-align: baseline;
	}
	& tr.table-wrapper td:first-child {
		border-radius: 12px 0 0 12px;
		white-space: nowrap;
	}
	& tr.table-wrapper td:last-child {
		border-radius: 0 12px 12px 0;
	}
	.text-wrap {
		white-space: wrap;
	}
}
</style>
