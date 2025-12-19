<template>
	<MyPage2>
		<MyCard flat :title="t('DETAILS')">
			<DetailPage :data="details" colWidth="240px">
				<!-- <template v-slot:Password="data">
          <div class="row items-end">
            <span>{{ passworkFormat(data.data.value) }}</span>
            <q-img
              :src="passwordEditIcon"
              fit="cover"
              width="24px"
              @click="passworkHandler(data.data)"
            />
          </div>
        </template> -->
			</DetailPage>
		</MyCard>
		<MyCard no-content-gap flat :title="tableTitle">
			<QTableStyle>
				<q-table
					:rows="database"
					:columns="columns"
					row-key="name"
					flat
					hide-pagination
					v-model:pagination="pagination"
					:rows-per-page-label="$t('RECORDS_PER_PAGE')"
				>
					<template v-slot:body-cell-password="props">
						<q-td :props="props">
							<PasswordToggle
								style="min-width: 120px"
								class="no-wrap"
								:value="props.value"
							></PasswordToggle>
						</q-td>
					</template>
					<template v-slot:body-cell-name="props">
						<q-td :props="props">
							<div class="row wrap middleware-chip-container">
								<q-chip
									class="middleware-chip-warpper"
									v-for="item in props.value"
									:key="item"
									outline
									square
									dense
									color="separator"
								>
									<span class="text-ink-1">{{ item }}</span>
								</q-chip>
							</div>
						</q-td>
					</template>
					<template v-slot:no-data>
						<div class="row justify-center full-width q-mt-lg">
							<Empty></Empty>
						</div>
					</template>
				</q-table>
			</QTableStyle>
		</MyCard>
	</MyPage2>
	<q-dialog v-model="visible" persistent>
		<q-card style="width: 480px">
			<q-card-section class="row items-center q-pb-none">
				<div class="text-h6">{{ t('CHANGE_PASSWORD') }}</div>
				<q-space />
				<q-btn icon="close" flat round dense v-close-popup />
			</q-card-section>
			<q-card-section class="q-mt-md">
				<q-form>
					<q-input
						v-model="user.name"
						type="text"
						outlined
						dense
						disable
						clearable
						:rules="[
							(val) => (val && val.length > 0) || 'Please type something'
						]"
					>
						<template v-slot:before>
							<div class="form-before">Admin</div>
						</template>
					</q-input>
					<q-input
						v-model="user.password"
						:type="isPwd ? 'password' : 'text'"
						outlined
						dense
						clearable
						:rules="[
							(val) => (val && val.length > 0) || t('PASSWORD_EMPTY_DESC'),
							(val) => PATTERN_PASSWORD.test(val) || t('PASSWORD_INVALID_DESC')
						]"
					>
						<template v-slot:before>
							<div class="form-before">Password</div>
						</template>
						<template v-slot:append>
							<q-icon
								:name="isPwd ? 'visibility_off' : 'visibility'"
								class="cursor-pointer"
								@click="isPwd = !isPwd"
							/>
						</template>
					</q-input>
				</q-form>
			</q-card-section>
			<q-card-actions align="right">
				<q-btn :label="t('OK')" type="submit" color="primary" @click="submit" />
			</q-card-actions>
			<q-inner-loading :showing="loading2"> </q-inner-loading>
		</q-card>
	</q-dialog>
	<q-inner-loading :showing="loading"> </q-inner-loading>
</template>

<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import DetailPage from '@apps/control-panel-common/src/containers/DetailPage.vue';
import MyCard from '@apps/control-panel-common/src/components/MyCard2.vue';
import { useAppDetailStore } from '@apps/control-hub/src/stores/AppDetail';
const appDetailStore = useAppDetailStore();
const username = appDetailStore.user.username;
import {
	getMiddlewareAll,
	getMiddlewareList,
	updateMiddlewarePassword
} from '@apps/control-hub/src/network';
import {
	MiddlewareItem,
	MiddlewareType
} from '@apps/control-panel-common/src/network/middleware';
import Empty from '@apps/control-panel-common/src/components/Empty.vue';
import Empty2 from '@apps/control-panel-common/src/components/Empty2.vue';
import { useQuasar } from 'quasar';
import { PATTERN_PASSWORD } from '@apps/control-panel-common/src/utils/constants.js';
import { useRoute } from 'vue-router';
import QTableStyle from '@apps/control-panel-common/src/components/QTableStyle2.vue';
import MyPage2 from '@apps/control-panel-common/src/containers/MyPage2.vue';
import { useI18n } from 'vue-i18n';
import PasswordToggle from '@apps/control-panel-common/src/components/PasswordToggle.vue';
import { useMiddlewareStore } from '@apps/control-hub/stores/Middleware';
import { get } from 'lodash';

const middlewareStore = useMiddlewareStore();
const { t } = useI18n();
const route = useRoute();
const $q = useQuasar();
const visible = ref(false);
const user = reactive({
	name: '',
	password: ''
});
const isPwd = ref(true);
const loading = ref(false);
const loading2 = ref(false);
const columns: any = computed(() => {
	const { type }: Record<string, any> = route.params;
	const USERNAME_NAME = 'username';

	const nameLabelMap: Record<MiddlewareType, string> = {
		mongodb: t('DATABASE'),
		postgres: t('DATABASE'),
		mysql: t('DATABASE'),
		mariadb: t('DATABASE'),
		redis: t('DATABASE'),
		rabbitmq: t('VHOST'),
		minio: t('BUCKETS'),
		nats: t('SUBJECT'),
		elasticsearch: t('INDEX')
	};
	const nameLabel = nameLabelMap[type] || t('DATABASE');

	const appLabel = t('middleware.app');

	const userLabelMap: Record<MiddlewareType, string> = {
		mongodb: t('middleware.user_database'),
		postgres: t('middleware.user_database'),
		mysql: t('middleware.user_database'),
		mariadb: t('middleware.user_database'),
		redis: t('middleware.user_database'),
		rabbitmq: t('middleware.user_other'),
		minio: t('middleware.user_other'),
		nats: t('middleware.user_other'),
		elasticsearch: t('middleware.user_other')
	};
	const userLabel = userLabelMap[type] || t('middleware.user_database');

	let data = [
		{
			name: 'app',
			required: true,
			label: appLabel,
			align: 'left',
			field: 'app'
		},
		{
			name: USERNAME_NAME,
			align: 'left',
			label: userLabel,
			field: 'username',
			format: (val) => val || '-'
		},
		{
			name: 'name',
			label: nameLabel,
			field: 'name',
			align: 'left',
			style: {
				maxWidth: '60%'
			}
		},
		{ name: 'password', label: t('PASSWORD'), field: 'password', align: 'left' }
	];

	if (type === 'redis') {
		data = data.filter((item) => item.name !== USERNAME_NAME);
		data.splice(1, 0, {
			name: 'namespace',
			align: 'left',
			label: t('NAMESPACE'),
			field: 'namespace',
			format: (val) => val || '-'
		});
	}

	return data;
});
const pagination = ref({
	rowsNumber: 0
});
const passworkFormat = (value: string | number) => {
	return '*'.repeat(6);
};

const detailsFormat = (data: MiddlewareItem) => {
	if (!data) return [];
	const { type: middleware }: Record<string, any> = route.params;

	const baseConfig = [
		{
			name: t('CLUSTER'),
			value: 'default'
		},
		{
			name: t('NAMESPACE'),
			value: data.namespace
		},
		{
			name: t('PASSWORD'),
			value: data.password
		}
	];

	if (middleware === 'mongodb') {
		return [
			...baseConfig,
			{
				name: t('MONGOS'),
				value: get(data, 'mongos.endpoint') || get(data, 'proxy.endpoint')
			},
			{
				name: t('USER'),
				value: data.adminUser
			}
		];
	}

	if (middleware === 'redis') {
		return [
			...baseConfig,
			{
				name: t('PROXY'),
				value: get(data, 'redisProxy.endpoint') || get(data, 'proxy.endpoint')
			}
		];
	}

	return [
		...baseConfig,
		{
			name: t('HOST'),
			value: get(data, 'proxy.endpoint')
		},
		{
			name: t('USER'),
			value: data.adminUser
		}
	];
};

const details = computed(() => detailsFormat(currentData.value));

const currentData = ref();
const tableTitle = computed(() => {
	const { type }: Record<string, any> = route.params;
	const titleMap: Record<string, string> = {
		mongodb: t('DATABASE'),
		postgres: t('DATABASE'),
		postgresql: t('DATABASE'),
		mysql: t('DATABASE'),
		mariadb: t('DATABASE'),
		redis: t('DATABASE'),
		rabbitmq: t('VHOST'),
		minio: t('BUCKETS'),
		nats: t('SUBJECT'),
		elasticsearch: t('INDEX'),
		es: t('INDEX')
	};
	return titleMap[type] || t('DATABASE');
});

const database = computed(() => {
	const data = currentData.value?.databases;
	let databaseItem = [];
	if (currentData.value) {
		return data.map((item: any) => {
			databaseItem = item.databases || item.buckets || [];

			return {
				name: databaseItem.map((item: any) => item.name),
				username: item.username,
				app: item.name,
				password: item.password,
				namespace: item.namespace
			};
		});
	} else {
		return [];
	}
});

const submit = () => {
	const { type: middleware, namespace }: Record<string, any> = route.params;
	const params = {
		name: currentData.value.name,
		namespace: currentData.value.namespace,
		middleware: middleware,
		user: currentData.value.username,
		password: user.password
	};
	loading2.value = true;
	updateMiddlewarePassword(middleware, params)
		.then((res) => {
			$q.notify({
				type: 'positive',
				message: res.data.message
			});
			loading2.value = false;
			visible.value = false;
		})
		.catch((err) => {
			$q.notify({
				type: 'negative',
				message: err.message
			});
			loading2.value = false;
		});
};

const fetchData = () => {
	loading.value = true;
	const { type: middleware }: Record<string, any> = route.params;
	getMiddlewareAll()
		.then((databaseList) => {
			try {
				const target: any = middlewareStore.list.find(
					(item) => item.type === middleware
				);
				const databases = databaseList.data.data.filter(
					(child: any) => child.type === middleware
				);
				currentData.value = {
					...target,
					databases
				};
			} catch (error) {
				currentData.value = undefined;
			}
		})
		.finally(() => {
			loading.value = false;
		});
};

watch(
	() => route.fullPath,
	() => {
		fetchData();
	},
	{
		immediate: true
	}
);
</script>

<style lang="scss" scoped>
.my-menu-link {
	padding: 8px;
}
.my-menu-link-active {
	background-color: rgba(34, 111, 255, 0.12);
}
.form-before {
	width: 88px;
	font-size: 14px;
	color: #484848;
}
.middleware-chip-container {
	gap: 6px;
	.middleware-chip-warpper {
		border-radius: 4px;
		margin: 0px;
	}
}
</style>
