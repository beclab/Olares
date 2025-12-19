<template>
	<div class="youtube-search-container relative-position">
		<q-resize-observer @resize="onResize" />
		<q-card flat bordered class="my-card bg-white">
			<div class="q-gutter-y-lg">
				<q-card-section class="q-pa-none">
					<div class="row items-center no-wrap">
						<q-img
							:src="vaultPluginIcon"
							:ratio="1"
							width="24px"
							spinner-size="xs"
						/>
						<div class="col q-ml-sm">
							<div class="text-subtitle1 text-grey-10">LarePass</div>
						</div>

						<div class="col-auto">
							<q-img
								:src="Image80"
								height="20px"
								width="20px"
								spinner-color="primary"
								spinner-size="xs"
								style="cursor: pointer"
							/>
							<q-btn
								flat
								color="light-blue-6 q-pa-none my-btn-wrapper relative-position q-ml-md"
								no-caps
								class="bg-grey-1"
								style="min-height: 24px"
								@click="clickHandler"
							>
								<div class="row items-center">
									<span class="q-ml-xs text-subtitle3 text-grey-10"
										>Summarize Video</span
									>
									<MyLoading v-show="loading" class="q-ml-xs"></MyLoading>
								</div>
							</q-btn>
						</div>
					</div>
				</q-card-section>

				<template v-if="text">
					<div class="text-body2 text-grey-10">
						<div>{{ name }}</div>
						{{ text }}
					</div>

					<div class="row q-gutter-x-md">
						<q-btn
							outline
							color="light-blue-6 q-pa-none my-btn-wrapper"
							no-caps
							style="min-height: 24px"
						>
							<q-img
								:src="vaultPluginBtn1"
								:ratio="1"
								width="20px"
								spinner-size="xs"
							/>
							<span class="q-ml-xs text-body3">Continue in Chat</span>
						</q-btn>
						<q-btn
							outline
							color="light-blue-6 q-pa-none my-btn-wrapper"
							no-caps
							style="min-height: 24px"
						>
							<q-icon name="sym_r_share" size="18px" />
							<span class="q-ml-xs text-body3">Add to Memo</span>
						</q-btn>
					</div>
					<q-separator
						style="margin-left: -16px; margin-right: -16px"
						class="bg-grey-2"
					/>
					<MyTimeline></MyTimeline>
				</template>
			</div>
		</q-card>
	</div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useQuasar } from 'quasar';
import vaultPluginIcon from 'assets/images/vault-plugin.svg';
import vaultPluginBtn1 from 'assets/images/vault-plugin-btn-1.svg';
import Image80 from 'assets/images/Image80.svg';
import MyTimeline from './components/MyTimeline.vue';
import MyLoading from './components/MyLoading.vue';
const $q: any = useQuasar();

const loading = ref(true);
const text = ref('');
const name = ref('');

let locker: any = null;
const clickHandler = () => {
	if (text.value.length > 300) {
		clearInterval(locker);
		return;
	}
	locker = setInterval(() => {
		text.value = `${text.value}-${Math.random().toString()}`;
	}, 100);
};

const onResize = async (size: { width: number; height: number }) => {
	const data = await $q.bex.send('webos.youtube.search.resize', {
		height: size.height // So it knows to make it bigger / smaller
	});

	name.value = data.data.name;
};
</script>

<style lang="scss" scoped>
.youtube-search-container {
}
.my-card {
	padding: 16px 16px 20px 16px;
	border-radius: 12px;
	border: 1px solid $separator;
	.my-tabs-active-class {
		color: $grey-10;
		background: white;
	}
	.my-tab-panels-wrapper {
		border-radius: 0px;
	}
	.my-tab-panel-wrapper {
		padding: 20px 0 0 0;
	}
}
</style>
<style lang="scss" scoped>
.my-btn-wrapper {
	padding: 2px 8px;
	border-radius: 4px;
}
</style>
