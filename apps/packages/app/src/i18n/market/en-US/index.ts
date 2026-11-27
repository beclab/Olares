// This is just an example,
// so you can safely delete all default props below

export default {
	detail: {
		require_memory: 'Required memory',
		require_disk: 'Required disk',
		require_cpu: 'Required CPU',
		require_gpu: 'Required GPU',
		resource: 'Resources',
		accelerator: 'Accelerator',
		gpu_memory: 'VRAM',
		more_modes: '+{count} more',
		more_languages: '+ {count} more',
		na: 'N/A',
		custom: 'Custom',
		resource_insufficient_banner:
			'This compute mode does not meet the app’s resource requirements: {reason}',
		no_matching_accelerator_banner:
			'No compatible compute resources are available in this Olares cluster',
		shortfall_cluster_cpu: 'Not enough CPU in this Olares cluster',
		shortfall_cluster_memory: 'Not enough memory in this Olares cluster',
		shortfall_cluster_disk: 'Not enough disk space in this Olares cluster',
		shortfall_cluster_gpu: 'Not enough VRAM in this Olares cluster',
		about_this_type: 'About this {type}',
		whats_new: 'What’s new',
		required_permissions: 'Required permissions',
		readme: 'Readme',
		information: 'Information',
		get_support: 'Get support',
		website: 'Website',
		app_version: 'App version',
		compatibility: 'Compatibility',
		platforms: 'Platforms',
		source_code: 'Source code',
		public: 'Public',
		legal: 'Legal',
		license: 'License',
		chart_version: 'Chart version',
		version_history: 'Version history',
		get_a_client: 'Get a client',
		dependency: 'Dependency',
		reference_app: 'Reference apps',
		see_all_version: 'See all versions',
		download: 'Download',
		no_version_history_desc: 'The app has no version history',
		dependency_not_installed: 'Dependency not installed',
		require_dependencies_for_full:
			'The following is required for full functionality:',
		reference_app_not_installed: 'Reference app not installed',
		need_reference_app_to_use:
			'This is a shared app. Install the corresponding reference app to use it.'
	},
	permission: {
		index: 'Permissions'
	},
	error: {
		index: 'Error',
		network_error: 'Network error. Try again later.',
		unknown_error: 'Unknown error',
		failed_get_user_role: 'Unable to get user role',
		only_be_installed_by_the_admin:
			'Only the cluster admin can install this app',
		not_admin_role_install_middleware:
			'This is a middleware component. Contact your Olares cluster admin to install it.',
		not_admin_role_install_cluster_app:
			'This is a shared app. Your Olares cluster admin must install it before you can use it.',
		failed_to_get_os_version: 'Unable to get the Olares version',
		app_is_not_compatible_terminus_os: 'Incompatible with this Olares version',
		failed_to_get_user_resource: 'Unable to get user resources',
		user_not_enough_cpu: 'Insufficient CPU in your quota',
		user_not_enough_memory: 'Insufficient memory in your quota',
		failed_to_get_system_resource: 'Unable to get system resources',
		need_to_install_dependent_app_first: 'Install the dependent app first',
		terminus_not_enough_cpu: 'Insufficient CPU in the Olares cluster',
		terminus_not_enough_memory: 'Insufficient memory in the Olares cluster',
		terminus_not_enough_disk: 'Insufficient disk space in the Olares cluster',
		terminus_not_enough_gpu: 'Insufficient GPU in the Olares cluster',
		operation_preform_failure: 'Operation failed',
		app_info_get_failure: 'Unable to get app information',
		cluster_not_support_platform:
			'This [app/recommend/model] does not support your Olares platform',
		app_install_conflict:
			'This app conflicts with installed apps. Uninstall the conflicting app and try again:',

		app_not_install_details: 'App name: {name}, version: {version}',
		middleware_not_install_details: 'Middleware: {name}, version: {version}',
		app_install_conflict_details: '{name}',
		failed_to_get_your_olares_platform_architecture:
			'Unable to get the Olares platform architecture',
		no_supportarch_specified_in_app_chart:
			'No supported architecture is specified in the app chart',
		source_conflict:
			'Source conflict: an app with the same name from a different source is already installed. Uninstall the existing app before installing this version.',
		the_dependent_app_must_be_installed_first:
			'The dependent app must be installed first. Contact your Olares cluster admin to install the required shared app.',
		csv2_install_blocked:
			'This app version has been deprecated and cannot be installed',
		csv2_upgrade_blocked:
			'This app version has been deprecated and cannot be upgraded',
		sharedv3_subuser_install_blocked:
			'This app cannot be installed. Contact an admin to install it.',
		sharedv3_subuser_operation_blocked:
			'Your account is not allowed to perform this action. Contact an admin.',
		template_install_blocked:
			'This is a template app and cannot be installed directly. Create an instance from it instead.'
	},
	compute: {
		select_mode_title: 'Choose a hardware accelerator',
		select_mode_desc:
			'This app supports multiple hardware accelerators. Once selected, you cannot change this after installation.',
		select_mode_placeholder: 'Select an accelerator',
		mode_status: {
			installable: 'Installable',
			insufficient_resources: 'Not enough resources',
			no_matching_node: 'No matching node'
		},
		mode_status_tip: {
			installable: 'This mode is available on your cluster',
			insufficient_resources:
				'A matching node exists, but it does not have enough free VRAM or GPUs',
			no_matching_node: 'No node in your cluster supports this compute mode'
		},
		select_binding_title: 'Choose GPU assignment for "{title}"',
		select_binding_desc: 'Choose the node and GPUs this app should run on',
		binding_unavailable_banner:
			'Your previous GPU selection is no longer available. Pick a different node or GPU.',
		validation_code: {
			aggregate_vram_insufficient:
				'The selected GPUs do not have enough combined VRAM. Add more GPUs or pick another node.',
			device_vram_insufficient:
				'This GPU does not have enough VRAM. Remove apps from this GPU to free up VRAM first.',
			device_memory_insufficient:
				'The node for this GPU does not have enough memory. Pick another GPU or node.',
			node_pressure:
				'Unable to assign to this accelerator. Scheduling the app to this node will cause overload. Please retry later or switch to another node.'
		},
		node_pressure_dimension:
			'{resource}: Total {total}, Used {used}, Required {needed}',
		pressure_resource: {
			memory: 'Memory',
			cpu: 'CPU',
			disk: 'Disk'
		},
		required_gpu: 'VRAM required',
		required_memory: 'Required memory',
		mode: 'Mode',
		support_multi_card: 'Multi-GPU supported on a single node',
		support_multi_nodes: 'Multi-GPU supported across nodes',
		support_type: 'Support type',
		node_status: {
			available: 'Available',
			not_available: 'Insufficient free VRAM',
			not_enough: 'Not enough total VRAM',
			not_match: 'Incompatible'
		},
		node_status_tip: {
			available: 'This node can run the app',
			not_available:
				'This node has enough total VRAM, but not enough free VRAM. Remove assigned apps to free up VRAM.',
			not_enough: 'This node’s total VRAM is below this app’s requirement',
			not_match: 'This node does not match the selected compute mode'
		},
		device_not_operable:
			'This GPU is locked, unhealthy, or unavailable to Olares',
		must_same_node:
			'This app does not support cross-node assignment. Clear the current selection to pick on another node.',
		vram_input_help: 'Available: {max}.',
		vram_required: 'Enter the VRAM allocation in {unit}',
		vram_not_number: 'Enter numbers only',
		vram_exceeds_available: 'Allocate no more than {max} to this GPU',
		vram_exceeds_remaining: 'Requested VRAM exceeds the remaining capacity',
		vram_below_required: 'Allocate at least {min} to this GPU',
		launch_title: 'Launch {title}',
		select_compute_cards: 'Select GPUs',
		manage_accelerator: 'Manage accelerator',
		refresh: 'Refresh',
		getting_status: 'Getting accelerator status…',
		launch: 'Launch',
		confirm_launch: 'Confirm launch',
		vram_allocation: 'VRAM allocation',
		sufficient_vram: 'Sufficient VRAM',
		insufficient_vram: 'Insufficient VRAM',
		cpu_label: 'CPU',
		memory_label: 'Memory',
		cores: 'Cores',
		single_card_chip: 'Single-GPU',
		multi_card_chip: 'Multi-GPU',
		multi_node_chip: 'Multi-node',
		apps_bound: 'Assigned apps',
		allocated: '{size} allocated',
		unbind: 'Remove',
		only_one_card: 'Only one GPU can be assigned',
		only_one_app_per_node:
			'Only one app per node is allowed for this resource type',
		card_insufficient_banner:
			'This GPU has enough total VRAM, but not enough free VRAM. Remove assigned apps to free up VRAM.',
		card_not_enough_banner:
			'This GPU’s total VRAM is below this app’s requirement. It cannot be assigned to this app.',
		unbind_title: 'Remove {app}?',
		unbind_desc:
			'Removing {app} from {card} stops the app and releases its GPU resources. To use it again, assign resources and launch it.',
		unbind_desc_all:
			'Removing all GPU assignments for {app} stops the app and releases its GPU resources. To use it again, assign resources and launch it.',
		dedicated_vram: 'Dedicated VRAM',
		shared_ram: 'Shared RAM',
		quota_for_current_app: 'Quota for current app',
		quota_available: '{size} available'
	},
	my: {
		all: 'All',
		olares: 'Olares',
		local: 'Local',
		unable_to_install_app: 'Unable to install app',
		update_all: 'Update all',
		available_updates: 'Available updates',
		everything_up_to_date: 'Everything is up to date',
		no_upload_chart_tips: 'You haven’t uploaded any custom app packages yet',
		no_installed_app_tips: 'You haven’t installed any apps yet',
		no_logs: 'No installation logs',
		sure_to_uninstall_the_app: 'Uninstall {title}?',
		upload_custom_chart: 'Upload custom app package',
		logs: 'Logs'
	},
	manage: 'Manage',
	updates: 'Updates',
	search_result_for: 'Search results for {content}',

	//temp
	discover_amazing_apps: 'Discover amazing apps',
	featured_app: 'Featured app',
	control_your_own_social_network: 'Control your own social network',
	decentralized_social_media: 'Decentralized social media',
	get_started: 'Get started',
	booster_your_software_development_productivity:
		'Boost your software development productivity',
	enjoy_coding_now: 'Enjoy coding now',
	diving_into_ai_image_generation: 'Dive into AI image generation',
	unleashing_your_creativity: 'Unleash your creativity',
	mastering_your_photo_with_a_personal_library:
		'Master your photos with a personal library',
	organize_your_memories: 'Organize your memories',
	nocode_solution_for_your_data_management:
		'No-code solution for your data management',
	build_databases_as_spreadsheets: 'Build databases as spreadsheets',

	top_app_in: 'Top apps in {category}',
	latest_app_in: 'Latest apps in {category}',
	recommend_app_in: 'Recommended apps in {category}',
	app_list: 'Applications',
	top_app_on_terminus: 'Top apps on Olares',
	latest_app_on_terminus: 'Latest apps on Olares',
	'Are you sure you want to delete this app chart from Local Sources?':
		'Remove this app package from local source?',
	'Enter a keyword to start searching': 'Search for apps',
	'Quickly find the apps you need': 'Search apps to install',
	no_find_app_in_search:
		'No apps match {keyword}\n\nYou can:\n- Check the spelling\n- Try different keywords\n- Browse by category',
	'Upload Chart': 'Upload app package',
	'Drop chart here to upload': 'Drag an app package here',
	'Uploading chart': 'Uploading app package {chart}',
	'Upload failed for chart': 'Unable to upload app package {chart}',
	'Processing app data': 'Processing app data',
	'App was successfully added to your local source. Do you want to install it now?':
		'{application} was added to your local source. Install it now?',
	'Template app was successfully added to your local source.':
		'{application} was added to your local source. Templates cannot be installed directly. Create an instance from it instead.',
	'An app named already exists in the local source. Do you want to update it with the one you just uploaded?':
		'An app named {application} already exists in the local source. Update it with the package you just uploaded?',
	'Incoming version:': 'Incoming version: ',
	'Installed version:': 'Installed version: ',
	'Install later': 'Install later',
	'Install now': 'Install now',
	'Cannot install app because a newer version already exists in your local source.':
		'Cannot install {application}. A newer version already exists in your local source.',
	'Adding Chart to Local Source': 'Adding app package {chart} to local source',
	'Failed to upload chart': 'Unable to upload app package {chart}',
	'Add app success': 'App added',
	'Download Raw Log': 'Download raw log',
	'Getting your app ready': 'Getting your app ready',
	'This can sometimes take a few moments. Thanks for your patience.':
		'This may take a while',
	'Entrance paused': 'Entrance paused',
	'Entrance to this application is currently paused. Please try resuming the app to re-start it.':
		'Resume this app to restart the entrance',
	Settings: 'Settings',
	'Market Source': 'Market source',
	'Choose a remote market source to retrieve application information.':
		'Select a remote market source to retrieve app information',
	'Source ID': 'Source ID',
	Description: 'Description',
	'Add Source': 'Add source',
	'Source Title': 'Source title',
	'Source URL': 'Source URL',
	'Change Source': 'Change source',
	'Are you sure you want to change the market sources? This might take some time to update.':
		'Change market sources? This may take a few minutes.',
	'Delete Source': 'Delete source',
	'Are you sure you want to delete this market source? This operation is permanent and cannot be undone.':
		'Delete this market source? This operation cannot be undone.',
	'Only alphanumeric characters are allowed.':
		'Only alphanumeric characters are allowed',
	'Source Title should be less than 10 characters':
		'Source title must be under 10 characters',
	'Source ID already exists. Please use a different name.':
		'Source ID {sourceId} already exists. Use a different name.',
	'Source URL already exists. Please use a different source URL':
		'Source URL already exists. Use a different source URL.',
	uninstall_options_desc: 'Select what else to remove',
	uninstall_shared_server_title: 'Uninstall shared server?',
	'Also uninstall the shared server (affects all users)':
		'Also uninstall the shared server (affects all users)',
	'Warning! Uninstalling the shared server will:':
		'This stops app access for all users and permanently deletes stored data for {appName}. This action cannot be undone.',
	'Cannot delete this source. Please uninstall all apps installed from this source, then try again.':
		'Cannot delete this source. Uninstall all apps installed from this source, then try again.',
	'Configure Environment Variables':
		'Configure environment variables for {title}',
	httpsRequired: 'The URL must start with https://',
	invalidUrlFormat: 'Invalid URL format',
	upgrade_app_version: `Update app from {currentVersion} to {targetVersion}`,
	'This value is set by a system environment variable':
		'This value comes from {envName}. To change it, go to Settings > Advanced > System environment variables.',
	'Clone App': 'Clone app',
	'Only lowercase letters and numbers are allowed; special characters and spaces are not permitted.':
		'Only lowercase letters and numbers are allowed. Special characters and spaces are not allowed.',
	'The length cannot exceed characters.':
		'Must be {length} characters or fewer',
	'Access to Files': 'Access to files',
	'Data, Cache and User directories': 'Data, cache, and user directories',
	'Data and Cache directories': 'Data and cache directories',
	'Data and User directories': 'Data and user directories',
	'Cache and User directories': 'Cache and user directories',
	'User directory': 'User directory',
	'Data directory': 'Data directory',
	'Cache directory': 'Cache directory',
	'Provide Public Entrance': 'Provide public entrance',
	'This app provides a publicly accessible entrance that does not require authentication.':
		'This app provides a public entrance that does not require authentication',
	"All traffic to this entrance may consume your reverse proxy's bandwidth.":
		'All traffic to this entrance may consume reverse proxy bandwidth',
	'No Visible Entrance': 'No visible entrance',
	'This is a background service with no user interface. It provides an API for other apps to interact with.':
		'This is a background service with no user interface. It provides an API for other apps to interact with.',
	'Shared App': 'Shared app',
	'This app is shared by all users in the same Olares cluster.':
		'This app is shared by all users in the same Olares cluster',
	'Administrator Only': 'Only cluster admins can install',
	'This app requires Olares administrator privileges to install.':
		'This app requires Olares cluster admin privileges to install',
	'Multiple Instances': 'Multiple instances',
	'This app can be cloned into multiple independent instances after installation.':
		'This app can be cloned into multiple independent instances after installation',
	'Using System Environment Variables': 'Use system environment variables',
	'This app retrieves its env values from the following system environment variables.':
		'This app uses values from these system environment variables:',
	'Connect to Other Apps': 'Connect to other apps',
	'This app requires permission to call following providers to extend its functionality':
		'This app needs permission to call the following providers:',
	'Using Middleware on Olares': 'Use middleware on Olares',
	'This app requires the following middleware.':
		'This app requires the following middleware:',
	'Expose Ports for Remote Access': 'Expose ports for remote access',
	'This app opens the following ports for remote access when VPN is enabled':
		'This app opens the following ports for remote access when VPN is enabled:',
	'New Application Title': 'New app name',
	'Please provide a unique title for your cloned application to avoid confusion with existing ones.':
		'Enter a unique name to avoid conflicts with existing apps',
	'Please provide a unique title for the app entrance on Desktop':
		'Enter a unique shortcut name for Desktop',
	'Desktop shortcut name for': 'Shortcut name for {name}',
	Clone: 'Clone',
	Instances: 'Instances',
	Create: 'Create',
	View: 'View',
	'{count} Instances created':
		'{count} instance created | {count} instances created',
	'Created from {name}': 'Created from {name}',
	'The source app is no longer available.':
		'The source app is no longer available',
	'From App': 'From app',
	'From Template': 'From template',
	'Create New Instance': 'Create new instance',
	'Upload response error': 'Upload response error',
	'Upload failed': 'Upload failed',
	'Also stop the shared server （affects all users)':
		'Also stop the shared server (affects all users)',
	'Are you sure you want to stop the shared app': 'Stop shared app {app}?',
	'Are you sure you want to resume the shared app "{app}" ?（affects all users）':
		'Resume shared app {app}? This affects all users.',
	'Do you want to buy app for': 'Buy {name} for {price} {unit}?',
	'Purchase App': 'Purchase app',
	Cancel: 'Cancel',
	Buy: 'Buy',
	'Restore purchase': 'Restore purchase',
	'This app has been removed, the operation is rejected.':
		'Cannot continue because this app has been removed',
	'NSFW settings': 'NSFW settings',
	'Blocking NSFW applications': 'Block NSFW apps',
	'No log records available': 'No logs',
	'get log records error': 'Unable to load logs',

	debug: {
		common: {
			appName: 'App name',
			version: 'Version',
			copied: 'Copied',
			copy: 'Copy',
			refresh: 'Refresh',
			export: 'Export',
			clear: 'Clear',
			all: 'All',
			noMatchedData: 'No matching data'
		},
		panel: {
			nav: {
				monitor: 'Status monitor',
				map: 'Map data',
				search: 'App search',
				labels: 'Labels',
				entryDiff: 'Entry diff'
			}
		},
		appSearch: {
			title: 'App search',
			keyword: 'Search keyword',
			placeholder: 'Enter app name, title, or description',
			found: 'Found {count} results',
			emptyKeyword: 'Enter keywords to search apps',
			noMatch: 'No matching app found',
			resultList: 'Search results',
			selectHint: 'Select a result to view details',
			weight: 'Weight'
		},
		jsonView: {
			selectKey: 'Select a key to view data',
			copyFailed: 'Unable to copy'
		},
		sidebar: {
			searchPlaceholder: 'Search key',
			noMatchedKey: 'No matching key'
		},
		mapViewer: {
			title: 'Map data viewer'
		},
		labels: {
			title: 'App labels viewer',
			totalWithLabels: '{count} apps with labels',
			groupStats: 'Label group stats',
			singleStats: 'Single label stats',
			appList: 'App list',
			selectApp: 'Select an app to view details'
		},
		statusFloat: {
			title: 'Status monitor',
			clickToStart: 'Select start to monitor',
			dragHint: 'Drag header to move. Total: {count}',
			openFullPage: 'Open full page'
		},
		statusMonitor: {
			title: 'App status monitor',
			stop: 'Stop',
			start: 'Start',
			import: 'Import',
			filterPlaceholder: 'Filter app name',
			clear: 'Clear',
			export: 'Export',
			importReadonlyMode: 'Imported snapshot mode (read-only)',
			fileLabel: 'File: ',
			exportTimeLabel: 'Export time: ',
			importTimeLabel: 'Import time: ',
			appCount: 'App count',
			callCount: 'Calls',
			total: 'Total',
			api: 'API',
			socket: 'Socket',
			local: 'Local',
			anomaly: 'Issues',
			rejected: 'Rejected',
			regression: 'Regression',
			clearFilter: 'Clear filter',
			monitoredApps: 'Monitored apps',
			updatesCount: '{count} updates',
			waitingStateChange: 'Waiting for state changes…',
			clickStartMonitor: 'Start monitoring',
			stateTransitionLog: 'State transition logs',
			viewAll: 'View all',
			onlySelectedApp: 'Only selected app',
			currentSelectedDetail: 'Current selected detail',
			oldTime: 'Old time:',
			newTime: 'New time:',
			diff: 'Diff:',
			collapse: 'Collapse',
			detail: 'Detail',
			noLogsForApp: 'No logs for selected app',
			noLogs: 'No logs',
			importFailed: 'Import failed',
			importMissingLogsArray: 'Import failed: missing logs array',
			importSuccess: '{count} logs imported. Snapshot mode is on.',
			importSnapshotFailed: 'Unable to import snapshot: {message}',
			checkFileFormat: 'Check the file format',
			noDetailData: 'No detail data available',
			serializeFailed: 'Unable to serialize detail data'
		},
		entryDiff: {
			title: 'app_entry / raw_data diff',
			exportJson: 'Export JSON',
			exportDoc: 'Export doc',
			total: 'Total',
			comparable: 'Comparable',
			mismatch: 'Mismatch',
			missingRaw: 'Missing raw_data',
			searchPlaceholder: 'Search key / app / diff path',
			allSources: 'All sources',
			currentSource: 'Current source',
			allFields: 'All fields',
			mismatchItems: 'Mismatch items',
			noMismatchDetected: 'No mismatch detected',
			detail: 'Detail',
			selectLeftItem: 'Select an item on the left to view details',
			missingAppEntry: 'Missing app_entry',
			fieldMismatch: 'Field mismatch',
			generatedAt: 'Generated at',
			filters: 'Filters',
			report: {
				overview: '## Overview',
				metricValue: '| Metric | Value |',
				appFullInfoTotal: 'appFullInfo total',
				currentViewCount: 'current view count',
				currentViewMismatchCount: 'current view mismatch count',
				currentViewMissingCount: 'current view missing count',
				bySource: '## By source',
				bySourceHeader: '| Source | Total | mismatch | missing |',
				byField: '## By mismatched field',
				byFieldHeader: '| Field | Count |',
				topDiffPaths: '## Top 30 diff paths',
				pathCountHeader: '| Path | Count |',
				fieldExamples: '## Field examples (1 per field)',
				noFieldExamples: 'No field examples under current filters.',
				fieldLabel: 'Field',
				relatedPaths: 'Related paths',
				appEntryBlock: '**app_entry block**',
				rawDataBlock: '**raw_data block**',
				itemsCurrentView: '## Item list (current view)',
				itemsHeader:
					'| key | source | app | status | fields | diffPathCount | diffPathsPreview |'
			}
		}
	},
	'getAppPaymentStatus result null': 'Unable to get payment status',
	'The backend is querying data, please wait.': 'Checking payment data…',
	'recoverAppPurchase result null': 'Unable to restore purchase',
	'getAppPurchase result null': 'Unable to get purchase information',
	'payment data empty': 'Payment data is missing',
	'Transaction submitted. Awaiting confirmation...':
		'Transaction submitted. Waiting for confirmation…',
	'Payment failed': 'Payment failed',
	'Transaction confirmed. Waiting for backend data synchronization...':
		'Transaction confirmed. Syncing data…',
	'Success, waiting for data synchronization...':
		'Payment confirmed. Syncing data…',
	'Failed to obtain txHash {txHash} and productId {productId}':
		'Unable to get transaction hash {txHash} and product ID {productId}',
	'Web3 instance not initialized (connect wallet first)':
		'Connect your wallet first',
	'MetaMask not detected (please install MetaMask)':
		'MetaMask not detected. Install MetaMask first.',
	'No accounts authorized': 'No wallet accounts authorized',
	'Unable to retrieve base fee': 'Unable to get base fee',
	'Invalid receiver address: {address}': 'Invalid receiver address: {address}',
	'Invalid transfer amount: {amount}': 'Invalid transfer amount: {amount}',
	'Encrypted order data is required': 'Order data is missing',
	'User cancelled transfer': 'Transfer canceled',
	'Transfer failed': 'Transfer failed',
	'Payment info cannot be empty': 'Payment information is missing',
	'Invalid price config: missing paid price information':
		'Price information is missing',
	'No supported token found for [symbol: {symbol}, chain: {chain}]':
		'No supported token found for {symbol} on {chain}',
	'Failed to get user account: wallet not connected':
		'Unable to get account. Wallet not connected.',
	'User has no balance for {symbol}': 'No {symbol} balance',
	'RSA public key is missing': 'Payment key is missing',
	'Failed to get encrypted order data': 'Unable to get encrypted order data',
	'Failed to get transaction hash after transfer':
		'Unable to get transaction hash after transfer',
	'paymentWithProduct error': 'Payment failed',
	'Failed to load app details, please try again.':
		'Unable to load app details. Please try again.',
	'No file selected.': 'No file selected',
	'Upload failed due to an unknown error.':
		'Unable to upload app package. Please try again.',
	'Upload has been canceled.': 'Upload canceled',
	'Upload failed. Please try again.':
		'Unable to upload app package. Please try again.',
	'App information is missing. Please install manually.':
		'App information is missing. Install it manually.'
};
