<template>
	<template v-if="loading">
		<tr v-for="item in 6" :key="item" class="table-wrapper">
			<td class="first-td">
				<q-skeleton style="width: 100%" height="26px" />
			</td>
			<td>
				<q-skeleton style="width: 100px" height="26px" />
			</td>
			<td>
				<q-skeleton style="width: 56px" height="26px" />
			</td>
		</tr>
	</template>
</template>
<script setup lang="ts">
interface Props {
	loading?: boolean;
}
withDefaults(defineProps<Props>(), {});
</script>
<style lang="scss" scoped>
.my-data-detail-table {
	border-collapse: separate;
	border-spacing: map-get($space-xs, x) map-get($space-md, x);
	width: 100%;
	margin: -12px -4px;
	.text-wrap {
		white-space: wrap;
	}
}

td:first-child {
	width: auto;
}

td:nth-child(2) {
	width: 100px;
}

td:last-child {
	width: 1%;
	white-space: nowrap;
}
</style>
