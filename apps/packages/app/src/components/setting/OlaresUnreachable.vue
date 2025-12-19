<template>
	<div class="olares-unreachable-bg q-pa-md">
		<div class="title text-suntitle3">
			{{ t('Olares unreachable. Try Search Olares if:') }}
		</div>
		<div class="row items-center justify-between bottom-bg">
			<div class="info">
				<div
					class="row text-body3 q-mt-sm"
					v-for="item in reminderList"
					:key="item"
				>
					<div
						class="row justify-center items-center"
						style="width: 8px; height: 16px"
					>
						<div
							style="width: 4px; height: 4px; border-radius: 2px"
							class="bg-ink-2"
						></div>
					</div>
					<div class="text-ink-2 q-ml-xs" style="width: calc(100% - 20px)">
						{{ item }}
					</div>
				</div>

				<div
					class="search text-body3 q-mt-md q-px-md row items-center justify-center"
					@click="searchOlares"
				>
					<div>
						<q-icon name="sym_r_search" size="16px" color="grey-10"></q-icon>
					</div>
					<div class="q-ml-xs">
						{{ t('Search Olares') }}
					</div>
				</div>
			</div>
			<div class="img">
				<img src="/larepass/olares_unreachable.svg" />
			</div>
		</div>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

const reminderList = ref([
	t('Olares is nearby but on a different network.'),
	t('Olares is on the same network and still starting up.')
]);

const emit = defineEmits(['searchOlares']);

const searchOlares = () => {
	emit('searchOlares');
};
</script>

<style scoped lang="scss">
.olares-unreachable-bg {
	background: #fff7f2;
	border: 1px solid $separator;
	border-radius: 12px;

	.title {
		color: $ink-1;
	}
	.bottom-bg {
		// background-color: red;

		.info {
			// background-color: green;
			max-width: calc(100% - 120px);

			.search {
				background: $yellow;
				border-radius: 8px;
				height: 32px;
				text-align: center;
				color: $grey-10;
				display: inline-flex;
			}
		}

		.img {
			width: 100px;
			height: 100px;
		}
	}
}
</style>
