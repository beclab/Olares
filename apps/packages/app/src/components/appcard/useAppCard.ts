import { getI18nValue, TRANSACTION_PAGE } from 'src/constant/constants';
import { getRawAppTitle, uninstalledApp } from 'src/constant/config';
import { useAppStore } from 'src/stores/market/appStore';
import { useMenuStore } from 'src/stores/market/menu';
import { useColor } from '@bytetrade/ui';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { computed } from 'vue';

export default function useAppCard(props: any) {
	const appStore = useAppStore();
	const menuStore = useMenuStore();
	const router = useRouter();
	const { locale } = useI18n();
	const appAggregation = computed(() => {
		return appStore.getAppAggregationInfo(props.appName, props.sourceId);
	});
	const { color: separatorColor } = useColor('separator');

	const clusterScopedApp = computed(
		() =>
			appAggregation.value?.app_full_info?.app_info?.app_entry?.options
				?.appScope?.clusterScoped ?? false
	);

	const appIcon = computed(
		() =>
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_icon ?? ''
	);

	const appTitle = computed(() => {
		return getRawAppTitle(
			locale.value,
			appAggregation.value?.app_status_latest?.status,
			appAggregation.value?.app_simple_latest?.app_simple_info
		);
	});

	const appDesc = computed(
		() =>
			getI18nValue(
				appAggregation.value?.app_simple_latest?.app_simple_info
					?.app_description,
				locale.value
			) ?? ''
	);

	const appVersion = computed(
		() =>
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_version ??
			''
	);

	const myAppVersion = computed(() => {
		if (uninstalledApp(appAggregation.value?.app_status_latest?.status)) {
			return '';
		}
		return appAggregation.value?.app_status_latest?.version ?? '';
	});

	const appFeaturedImage = computed(
		() =>
			appAggregation.value?.app_full_info?.app_info?.app_entry?.featuredImage ??
			appAggregation.value?.app_simple_latest?.app_simple_info?.app_icon ??
			''
	);

	function goAppDetails() {
		console.log('click app');
		if (props.disabled) {
			return;
		}
		router.push({
			name: TRANSACTION_PAGE.App,
			params: {
				appName: props.appName,
				sourceId: props.sourceId
			},
			query: {
				...router.currentRoute.value.query,
				menuItem: menuStore.currentItem
			}
		});
	}

	const sourceName = computed(() => {
		const source = appStore.sources.find((item) => item.id === props.sourceId);
		if (source) {
			return source.name;
		} else {
			return props.sourceId;
		}
	});

	return {
		separatorColor,
		appAggregation,
		clusterScopedApp,
		appIcon,
		appTitle,
		appDesc,
		appVersion,
		myAppVersion,
		appFeaturedImage,
		sourceName,
		goAppDetails
	};
}
