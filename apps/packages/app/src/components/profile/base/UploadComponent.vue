<template>
	<BtUploader
		class="bio-img"
		:size="5"
		fileName="image"
		accept=".jpg, .jpeg, .png, .gif, image/*"
		action="/images/upload/v1"
		@ok="ok"
		@loading="update"
		:width="width"
		:height="height"
		@fail="fail"
	>
		<q-img
			v-if="imgUrl || placeholder"
			:style="{ width: width, height: height }"
			class="bio-img"
			:src="imgUrl ? imgUrl : placeholder"
		>
			<template v-slot:loading>
				<base-inner-loading :width="width" :height="height" />
			</template>
			<template v-slot:default>
				<base-inner-loading v-if="loading" :width="width" :height="height" />
				<q-img v-else class="image-change" src="image_change.png" />
			</template>
		</q-img>
		<div
			v-else
			:style="{ width: width, height: height }"
			class="bio-img-none column justify-center items-center"
		>
			<base-inner-loading v-if="loading" :width="width" :height="height" />
			<q-icon v-else size="32px" color="ink-3" name="sym_r_imagesmode" />
		</div>
	</BtUploader>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';
import BaseInnerLoading from '@apps/profile/src/components/profile/base/BaseInnerLoading.vue';
import { ref } from 'vue';

defineProps({
	imgUrl: {
		type: String,
		require: true
	},
	placeholder: {
		type: String,
		require: false
	},
	width: {
		type: String,
		default: '72px'
	},
	height: {
		type: String,
		default: '72px'
	}
});
const emit = defineEmits(['update:imgUrl']);
const loading = ref(false);
const $q = useQuasar();

const ok = (response: {
	code: number;
	data: { imageUrl: string; size: any };
	message: string;
}) => {
	loading.value = false;
	console.log('ok ');
	if (response.code !== 200) {
		$q.notify(response.message);
		return;
	}

	emit('update:imgUrl', response.data.imageUrl);
};

const fail = (response: unknown) => {
	loading.value = false;
	BtNotify.show({
		type: NotifyDefinedType.FAILED,
		message: response
	});
};

const update = (status: boolean) => {
	console.log(status);
	loading.value = status;
};
</script>

<style scoped lang="scss">
.bio-img {
	border-radius: 8px;
	position: relative;
}

.image-change {
	position: absolute;
	top: calc(50% - 22px);
	left: calc(50% - 22px);
	width: 44px;
	height: 44px;
	background: transparent;
}

.bio-img-none {
	border-radius: 8px;
	background: $background-3;
	border: $separator-2;
}
</style>
