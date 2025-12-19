<template>
	<div class="card floating help">
		<div class="card-title">
			{{ $t('help.help') }}
		</div>

		<div class="card-content">
			<ul>
				<li v-for="item in platform" :key="item.key">
					<strong>{{ item.key }}</strong> - {{ item.des }}
				</li>
			</ul>
		</div>

		<div class="card-action">
			<button
				type="submit"
				@click="store.closeHovers()"
				:aria-label="$t('buttons.ok')"
				:title="$t('buttons.ok')"
			>
				{{ $t('buttons.ok') }}
			</button>
		</div>
	</div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDataStore } from '../../../stores/data';

interface platformType {
	key: string;
	des: string;
}

export default defineComponent({
	name: 'helpComponent',
	setup() {
		const store = useDataStore();
		const i18n = useI18n();
		const platform = ref<platformType[]>([]);
		const platformGroup = {
			Win: [
				{
					key: 'F1',
					des: i18n.t('help.f1')
				},
				{
					key: 'F2',
					des: i18n.t('help.f2')
				},
				{
					key: 'DEL',
					des: i18n.t('help.del')
				},
				{
					key: 'CTRL + C',
					des: i18n.t('help.ctrl.c')
				},
				{
					key: 'CTRL + X',
					des: i18n.t('help.ctrl.x')
				},
				{
					key: 'CTRL + V',
					des: i18n.t('help.ctrl.v')
				},
				{
					key: 'CTRL + S',
					des: i18n.t('help.ctrl.s')
				},
				{
					key: 'CTRL + Click',
					des: i18n.t('help.ctrl.click')
				},
				{
					key: 'Click',
					des: i18n.t('help.click')
				},
				{
					key: 'Double click',
					des: i18n.t('help.doubleClick')
				}
			],
			Mac: [
				{
					key: 'F1',
					des: i18n.t('help.f1')
				},
				{
					key: 'F2',
					des: i18n.t('help.f2')
				},
				{
					key: 'DEL',
					des: i18n.t('help.del')
				},
				{
					key: 'Command + C',
					des: i18n.t('help.ctrl.c')
				},
				{
					key: 'Command + X',
					des: i18n.t('help.ctrl.x')
				},
				{
					key: 'Command + V',
					des: i18n.t('help.ctrl.v')
				},
				{
					key: 'Command + S',
					des: i18n.t('help.ctrl.s')
				},
				{
					key: 'Command + Click',
					des: i18n.t('help.ctrl.click')
				},
				{
					key: 'Click',
					des: i18n.t('help.click')
				},
				{
					key: 'Double click',
					des: i18n.t('help.doubleClick')
				}
			]
		};

		onMounted(() => {
			let isWin = navigator.platform.indexOf('Win') === 0;
			if (isWin) {
				platform.value = platformGroup.Win;
			} else {
				platform.value = platformGroup.Mac;
			}
		});

		return {
			store,
			platform
		};
	}
});
</script>
