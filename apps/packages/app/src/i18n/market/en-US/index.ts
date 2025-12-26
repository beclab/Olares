// This is just an example,
// so you can safely delete all default props below

export default {
	detail: {
		require_memory: 'Required memory',
		require_disk: 'Required disk',
		require_cpu: 'Required CPU',
		require_gpu: 'Required GPU',
		about_this_type: 'About this {type}',
		whats_new: "What's new",
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
		no_version_history_desc: 'The application has no version history',
		dependency_not_installed: 'Dependency not installed',
		require_dependencies_for_full:
			'This application may require the following dependencies for full functionality.',
		reference_app_not_installed: 'Reference app not installed',
		need_reference_app_to_use:
			'This is a shared app, you will need the corresponding Reference app to use it.'
	},
	permission: {
		index: 'Permission'
	},
	error: {
		index: 'error',
		network_error: 'Network error. Please try again later.',
		unknown_error: 'Unknown error',
		failed_get_user_role: 'Failed to get user role',
		only_be_installed_by_the_admin: 'This app can only be installed by Admin',
		not_admin_role_install_middleware:
			'This is a middleware component. Contact your Olares Admin to install.',
		not_admin_role_install_cluster_app:
			'This is a shared app. It must be installed by your Olares Admin before you can use it.',
		failed_to_get_os_version: 'Failed to get Olares version',
		app_is_not_compatible_terminus_os: 'Incompatible with your Olares',
		failed_to_get_user_resource: 'Failed to get user resource',
		user_not_enough_cpu: 'Insufficient CPU on your quota',
		user_not_enough_memory: 'Insufficient memory on your quota',
		failed_to_get_system_resource: 'Failed to get system resource',
		need_to_install_dependent_app_first: 'Need to install dependent app first',
		terminus_not_enough_cpu: 'Insufficient CPU on the Olares cluster',
		terminus_not_enough_memory: 'Insufficient memory on the Olares cluster',
		terminus_not_enough_disk: 'Insufficient disk on the Olares cluster',
		terminus_not_enough_gpu: 'Insufficient GPU on the Olares cluster',
		operation_preform_failure: 'Operation failed',
		app_info_get_failure: 'Failed to retrieve app information',
		cluster_not_support_platform:
			'This [app/recommend/model] does not support your Olares platform',
		app_install_conflict:
			'This app conflicts with your installed apps. Please uninstall the following apps before proceeding with this installation.',

		app_not_install_details: 'App Name: {name}, Version: {version}',
		middleware_not_install_details: 'Middleware: {name}, Version: {version}',
		app_install_conflict_details: '{name}',
		failed_to_get_your_olares_platform_architecture:
			'Failed to get your Olares platform architecture',
		no_supportarch_specified_in_app_chart:
			'No supportArch specified in app chart',
		source_conflict:
			'Source Conflict: an app with the same name from a different source is already installed. Please uninstall the existing app before installing this version',
		the_dependent_app_must_be_installed_first:
			'The dependent app must be installed first. Please contact your Olares Admin to install the required shared app before proceeding. '
	},
	my: {
		all: 'All',
		olares: 'Olares',
		local: 'Local',
		unable_to_install_app: 'Unable to install app',
		update_all: 'Update all',
		available_updates: 'Available updates',
		everything_up_to_date: 'Congratulations! Everything is up to date.',
		no_upload_chart_tips: "You haven't uploaded any custom charts yet",
		no_installed_app_tips: 'You haven’t installed anything yet',
		no_logs: 'No installation logs available',
		sure_to_uninstall_the_app: "Are you sure to uninstall '{title}'？",
		upload_custom_chart: 'Upload custom chart',
		logs: 'Logs'
	},
	manage: 'Manage',
	updates: 'Updates',
	search_result_for: 'Search Results For "{content}"',

	//temp
	discover_amazing_apps: 'Discover amazing apps',
	featured_app: 'FEATURED APP',
	control_your_own_social_network: 'Control your own social network.',
	decentralized_social_media: 'Decentralized social media',
	get_started: 'GET STARTED',
	booster_your_software_development_productivity:
		'Boost your software development productivity.',
	enjoy_coding_now: 'Enjoy coding now',
	diving_into_ai_image_generation: 'Dive into AI image generation.',
	unleashing_your_creativity: 'Unleash your creativity!',
	mastering_your_photo_with_a_personal_library:
		'Master your photos with a personal library.',
	organize_your_memories: 'Organize your memories',
	nocode_solution_for_your_data_management:
		'No-code solution for your data management.',
	build_databases_as_spreadsheets: 'Build databases as spreadsheets',

	top_app_in: 'Top apps in {category}',
	latest_app_in: 'Latest apps in {category}',
	recommend_app_in: 'Recommend apps in {category}',
	app_list: 'Applications',
	top_app_on_terminus: 'Top apps on Olares',
	latest_app_on_terminus: 'Latest apps on Olares',
	'Are you sure you want to delete this app chart from Local Sources?':
		'Are you sure you want to delete this app chart from Local Sources?',
	'Enter a keyword to start searching': 'Enter a keyword to start searching',
	'Quickly find the apps you need': 'Quickly find the apps you need',
	no_find_app_in_search:
		'No apps match "{keyword}"\n\nYou can try:\n- Check the spelling.\n- Try different keywords.\n- Browse by category.',
	'Upload Chart': 'Upload chart',
	'Uploading chart': 'Uploading chart "{chart}"',
	'Upload failed for chart': 'Upload failed for chart "{chart}" ',
	'Processing app data': 'Processing app data...',
	'App was successfully added to your local source. Do you want to install it now?':
		'App "{application}" was successfully added to your local source. Do you want to install it now?',
	'An app named already exists in the local source. Do you want to update it with the one you just uploaded?':
		'An app named {application} already exists in the local source. Do you want to update it with the one you just uploaded?',
	'Incoming version:': 'Incoming version:',
	'Installed version:': 'Installed version:',
	'Install later': 'Install later',
	'Install now': 'Install now',
	'Cannot install app because a newer version already exists in your local source.': `Cannot install  {application} because a newer version already exists in your local source.`,
	'Adding Chart to Local Source': 'Adding chart "{chart}" to local source',
	'Failed to upload chart': 'Failed to upload chart {chart}',
	'Add app success': 'Add app success',
	'Download Raw Log': 'Download Raw Log',
	'Getting your app ready': 'Getting your app ready',
	'This can sometimes take a few moments. Thanks for your patience.':
		'This can sometimes take a few moments. Thanks for your patience.',
	'Entrance paused': 'Entrance paused',
	'Entrance to this application is currently paused. Please try resuming the app to re-start it.':
		'Entrance to this application is currently paused. Please try resuming the app to re-start it.',
	Settings: 'Settings',
	'Market Source': 'Market Source',
	'Choose a remote market source to retrieve application information.':
		'Choose a remote market source to retrieve application information.',
	'Source ID': 'Source ID',
	Description: 'Description',
	'Add Source': 'Add Source',
	'Source Title': 'Source Title',
	'Source URL': 'Source URL',
	'Change Source': 'Change Source',
	'Are you sure you want to change the market sources? This might take some time to update.':
		'Are you sure you want to change the market sources? This might take some time to update.',
	'Delete Source': 'Delete Source',
	'Are you sure you want to delete this market source? This operation is permanent and cannot be undone.':
		'Are you sure you want to delete this market source? This operation is permanent and cannot be undone.',
	'Only alphanumeric characters are allowed.':
		'Only alphanumeric characters are allowed.',
	'Source Title should be less than 10 characters':
		'Source Title should be less than 10 characters',
	'Source ID already exists. Please use a different name.':
		"Source ID '{sourceId}' already exists. Please use a different name.",
	'Source URL already exists. Please use a different source URL':
		'Source URL already exists. Please use a different source URL',
	'Also uninstall the shared server (affects all users)':
		'Also uninstall the shared server (affects all users)',
	'Warning! Uninstalling the shared server will:':
		'Warning! Uninstalling the shared server will:\n• Terminate the app access for all users immediately\n• Permanently delete all stored user data\n\nThis action cannot be undone. Please confirm that you acknowledge these consequences and wish to proceed with a full uninstallation of {appName}.',
	'Cannot delete this source. Please uninstall all apps installed from this source, then try again.':
		'Cannot delete this source. Please uninstall all apps installed from this source, then try again.',
	'Configure Environment Variables': 'Configure Environment Variables',
	httpsRequired: 'The URL must start with https://',
	invalidUrlFormat: 'Invalid URL format',
	upgrade_app_version: `Upgrade app version from {currentVersion} to {targetVersion}`,
	'This value is set by a system environment variable':
		'This value is set by a system environment variable. To change it, please go to Settings > Developer > System Environment and change [{envName}]',
	'Clone App': 'Clone app',
	'Only lowercase letters and numbers are allowed; special characters and spaces are not permitted.':
		'Only lowercase letters and numbers are allowed; special characters and spaces are not permitted.',
	'The length cannot exceed characters.':
		'The length cannot exceed {length} characters.',
	'Removal of the installation package is prohibited as there are cloned versions of the current application.':
		'Removal of the installation package is prohibited as there are cloned versions of the current application.',
	'Access to Files': 'Access to Files',
	'Data, Cache and User directories': 'Data, Cache and User directories',
	'Data and Cache directories': 'Data and Cache directories',
	'Data and User directories': 'Data and User directories',
	'Cache and User directories': 'Cache and User directories',
	'User directory': 'User directory',
	'Data directory': 'Data directory',
	'Cache directory': 'Cache directory',
	'Provide Public Entrance': 'Provide Public Entrance',
	'This app provides an publicly accessible entrance that does not require authentication.':
		'This app provides an publicly accessible entrance that does not require authentication.',
	"All traffic to this entrance may consume your reverse proxy's bandwidth.":
		"All traffic to this entrance may consume your reverse proxy's bandwidth.",
	'No Visible Entrance': 'No Visible Entrance',
	'This is a background service with no user interface. It provides an API for other apps to interact with.':
		'This is a background service with no user interface. It provides an API for other apps to interact with.',
	'Shared App': 'Shared App',
	'This app is shared by all users in the same Olares cluster.':
		'This app is shared by all users in the same Olares cluster.',
	'Administrator Only': 'Administrator Only',
	'This app requires Olares administrator privileges to install.':
		'This app requires Olares administrator privileges to install.',
	'Multiple Instances': 'Multiple Instances',
	'This app can be cloned into multiple independent instances after installation.':
		'This app can be cloned into multiple independent instances after installation.',
	'Using System Environment Variables': 'Using System Environment Variables',
	'This app retrieves its env values from the following system environment variables.':
		'This app retrieves its env values from the following system environment variables.',
	'Connect to Other Apps': 'Connect to Other Apps',
	'This app requires permission to call following providers to extend its functionality':
		'This app requires permission to call following providers to extend its functionality',
	'Using Middleware on Olares': 'Using Middleware on Olares',
	'This app requires the following middleware.':
		'This app requires the following middleware.',
	'Expose Ports for Remote Access': 'Expose Ports for Remote Access',
	'This app opens the following ports for remote access when VPN is enabled':
		'This app opens the following ports for remote access when VPN is enabled',
	'New Application Title': 'New app title',
	'Please provide a unique title for your cloned application to avoid confusion with existing ones.':
		'Please provide a unique title for your cloned application to avoid confusion with existing ones.',
	'Please provide a unique title for the app entrance on Desktop':
		'Please provide a unique title for the app entrance on Desktop',
	'Desktop shortcut name for': 'Desktop shortcut name for {name}',
	Clone: 'Clone',
	'Also stop the shared server （affects all users)':
		'Also stop the shared server (affects all users)',
	'Are you sure you want to stop the shared app':
		'Are you sure you want to stop the shared app "{app}"?',
	'Are you sure you want to resume the shared app "{app}" ?（affects all users）':
		'Are you sure you want to resume the shared app "{app}"? (affects all users)'
};
