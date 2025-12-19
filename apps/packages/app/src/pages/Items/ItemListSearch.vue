<template>
	<div class="searchWrap">
		<q-input
			class="searchInput text-yellow-7"
			v-model="filterInput"
			debounce="500"
			borderless
			dense
			placeholder="Search"
			@update:model-value="search"
		>
			<template v-slot:prepend>
				<q-icon name="sym_r_search" />
			</template>
			<template v-slot:append>
				<q-icon class="cursor-pointer" name="sym_r_close" @click="closeSearch">
					<q-tooltip>{{ t('buttons.close') }}</q-tooltip>
				</q-icon>
			</template>
		</q-input>
	</div>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const emits = defineEmits(['search', 'closeSearch']);

const filterInput = ref('');

const { t } = useI18n();

async function search() {
	emits('search', filterInput.value);
}

function closeSearch() {
	emits('closeSearch');
	if (filterInput?.value) {
		filterInput.value = '';
	}
}
</script>

<style lang="scss" scoped>
.searchWrap {
	width: 100%;
	height: 56px;
	line-height: 56px;
	text-align: center;

	.searchInput {
		padding: 0 8px;
		border: 1px solid $blue;
		border-radius: 10px;
		margin: 8px 16px;
		display: inline-block;
		display: flex;
		align-items: center;
		justify-content: center;
	}
}
</style>
