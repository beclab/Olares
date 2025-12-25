<template>
	<BtUploader
		class="bio-img"
		:size="5"
		fileName="image"
		accept=".jpg, .jpeg, .png, .gif, image/*"
		action="/images/upload/v1"
		@ok="ok"
		@loading="update"
		width="88px"
		height="88px"
		@fail="fail"
	>
		<q-img v-if="imgUrl" class="bio-img" :src="imgUrl">
			<template v-slot:loading>
				<base-inner-loading width="88" height="88" />
			</template>
		</q-img>
		<div v-else class="bio-img-none column justify-center items-center">
			<base-inner-loading v-if="loading" width="88" height="88" />
			<q-icon v-else size="32px" color="ink-3" name="sym_r_imagesmode" />
		</div>
	</BtUploader>
</template>

<script lang="ts" setup>
import { useQuasar } from 'quasar';
import BaseInnerLoading from '@apps/profile/src/components/profile/base/BaseInnerLoading.vue';
import { ref } from 'vue';
import { BtNotify, NotifyDefinedType } from '@bytetrade/ui';

defineProps({
	imgUrl: String
});

const $q = useQuasar();
const emit = defineEmits(['update:imgUrl']);
const loading = ref();

const ok = (response: {
	code: number;
	data: { imageUrl: string; size: any };
	message: string;
}) => {
	console.log('ok ');
	console.log(response.data);

	if (response.code !== 200) {
		$q.notify(response.message);
		return;
	}

	emit('update:imgUrl', response.data.imageUrl);
};

const fail = (response: unknown) => {
	console.log('fail', response);
	BtNotify.show({
		type: NotifyDefinedType.FAILED,
		message: response
	});
};

const update = (status: boolean) => {
	loading.value = status;
};
</script>

<style scoped lang="scss">
.bio-img {
	width: 88px;
	height: 88px;
	border-radius: 8px;
}

.bio-img-background {
	width: 88px;
	height: 48px;
}

.bio-img-none {
	width: 88px;
	height: 88px;
	border-radius: 8px;
	background: $background-3;
	border: $separator-2;
}
</style>
