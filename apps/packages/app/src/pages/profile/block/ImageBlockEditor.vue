<template>
	<div class="column" v-if="blockRef">
		<block-nickname
			class="q-mt-sm"
			:type="blockRef.type"
			v-model:text="blockRef.nickName"
		/>
		<div class="block-desc q-mt-lg">{{ t('blocks.feature_an_image') }}</div>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.add_headline') }}
		</div>
		<edit-view
			v-if="userStore.user"
			class="label-width q-mt-xs"
			:placeholder="t('blocks.image_headline')"
			v-model="blockRef.title"
		/>
		<div class="text-subtitle1 full-width q-mt-lg">
			{{ t('blocks.add_description') }}
		</div>
		<edit-view
			v-if="userStore.user"
			height="96px"
			class="label-width q-mt-xs"
			:placeholder="t('blocks.image_description')"
			v-model="blockRef.description"
		/>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.aspect_ratio') }}
		</div>
		<div class="checkbox-grid row">
			<template v-for="item in switchArray" :key="item.title">
				<check-box-component
					:label="item.title"
					:model-value="blockRef.ratio === item.value"
					@update:model-value="onRatioUpdate(item.value)"
				/>
			</template>
		</div>
		<div class="edit-label full-width q-mt-lg">
			{{ t('blocks.add_image') }}
		</div>
		<div class="image-border column q-pa-lg q-mt-md">
			<q-input
				dense
				borderless
				v-model="blockRef.link"
				:placeholder="t('blocks.image_link') + '(' + t('base.optional') + ')'"
			/>
			<upload-component
				class="q-mt-md"
				v-model:img-url="blockRef.img"
				width="100%"
				height="220px"
			/>
		</div>
	</div>
</template>

<script lang="ts" setup>
import CheckBoxComponent from '@apps/profile/src/components/profile/base/CheckBoxComponent.vue';
import UploadComponent from '@apps/profile/src/components/profile/base/UploadComponent.vue';
import BlockNickname from '@apps/profile/src/components/profile/block/BlockNickname.vue';
import EditView from '@apps/profile/src/components/profile/base/EditView.vue';
import { useUserStore } from '@apps/profile/src/stores/profileUser';
import { RATIO_TYPE } from '@apps/profile/src/types/User';
import { ref, onMounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

const userStore = useUserStore();
const route = useRoute();
const blockRef = ref();
const { t } = useI18n();

const switchArray = [
	{
		title: t('blocks.auto'),
		value: RATIO_TYPE.AUTO
	},
	{
		title: '1:1',
		value: RATIO_TYPE.RATIO_1_1
	},
	{
		title: '3:2',
		value: RATIO_TYPE.RATIO_3_2
	},
	{
		title: '16:9',
		value: RATIO_TYPE.RATIO_16_9
	},
	{
		title: '3:1',
		value: RATIO_TYPE.RATIO_3_1
	}
];

onMounted(() => {
	if (route.params.id && userStore.user) {
		const block = userStore.user.block.data.find((item) => {
			return item.id === route.params.id;
		});
		console.log(block);
		if (block) {
			blockRef.value = block;
		}
	}
});

const onRatioUpdate = (value: string) => {
	if (blockRef.value) {
		blockRef.value.ratio = value;
	}
};
</script>

<style scoped lang="scss">
.checkbox-grid {
	width: 100%;
	display: grid;
	align-items: start;
	justify-items: start;
	justify-content: center;
	grid-row-gap: 32px;
	grid-template-columns: repeat(5, minmax(0, 1fr));
}
.image-border {
	width: 100%;
	border-radius: 8px;
	border: 1px solid $input-stroke;
}
</style>
