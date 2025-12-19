<template>
	<div class="relative-position" style="overflow: hidden">
		<MenuHeader
			style="position: absolute; top: 0"
			class="text-left"
		></MenuHeader>
		<MyContentPage>
			<router-view></router-view>
		</MyContentPage>
	</div>
</template>

<script lang="ts" setup>
import MenuHeader from '@apps/control-hub/src/layouts/MenuHeader.vue';
import MyContentPage from '@apps/control-hub/src/components/MyContentPage.vue';
import { useMiddlewareStore } from '@apps/control-hub/stores/Middleware';
import { useRoute, useRouter } from 'vue-router';
import { computed, watch } from 'vue';
import {
	updateKey,
	updateKeyFirstOption
} from '@apps/control-hub/layouts/breadcrumbs';
import { useQuasar } from 'quasar';
import DialogConfirm from '@apps/control-hub/src/components/DialogConfirm.vue';
import { useI18n } from 'vue-i18n';
const route = useRoute();
const router = useRouter();
const $q = useQuasar();
const { t } = useI18n();
const useMiddleware = useMiddlewareStore();

const middlewareTypes = computed(() =>
	useMiddleware.list.map((item) => item.type)
);

watch(
	() => middlewareTypes.value,
	(newValue, oldValue) => {
		const { type }: Record<string, any> = route.params;

		const rest = oldValue.filter((item) => !newValue.includes(item));

		if (rest && rest.includes(type)) {
			$q.dialog({
				component: DialogConfirm,
				componentProps: {
					title: t('middleware.type_changed_title') + ` (${type})`,
					message: t('middleware.type_changed_message')
				}
			}).onOk(() => {
				updateKeyFirstOption();
				router.replace(`/`);
			});
		}
	}
);
</script>

<style lang="scss" scoped></style>
