// This is just an example,
// so you can safely delete all default props below

export default {
	detail: {
		require_memory: '所需内存',
		require_disk: '所需磁盘空间',
		require_cpu: '所需 CPU',
		require_gpu: '所需 GPU',
		about_this_type: '关于此{type}',
		whats_new: '新功能',
		required_permissions: '所需权限',
		readme: 'Readme',
		information: '信息',
		get_support: '获取支持',
		website: '网站',
		app_version: '应用版本',
		compatibility: '兼容性',
		platforms: '平台',
		source_code: '源代码',
		public: '公开',
		legal: '法律',
		license: '许可证',
		chart_version: 'Chart 版本',
		version_history: '版本历史',
		get_a_client: '获取客户端',
		dependency: '依赖项',
		reference_app: '授权应用',
		see_all_version: '查看所有版本',
		download: '下载',
		no_version_history_desc: '该应用没有版本历史记录',
		dependency_not_installed: '依赖应用未安装',
		require_dependencies_for_full:
			'本应用程序可能需要以下依赖项才能实现全部功能。',
		reference_app_not_installed: '授权应用未安装',
		need_reference_app_to_use:
			'该应用为共享应用，您需要安装对应的授权应用来使用它。'
	},
	permission: {
		index: '权限',
		files: '文件',
		files_not_store_label: '此应用不会在 Olares 上存储任何文件',
		files_access_user_data_label: '访问用户数据',
		files_access_user_data_desc: '允许此应用读取和写入以下目录中的文件：',
		file_store_data_folder_label: '访问 Data 文件夹',
		file_store_data_folder_desc: '允许此应用在特定应用的目录中持久化存储文件',
		file_store_cache_folder_label: '访问 Cache 文件夹',
		file_store_cache_folder_desc: '允许此应用在特定应用目录中存储缓存文件',
		internet: '互联网',
		internet_label: '安装和运行期间需要完全网络访问权限',
		internet_desc: '允许此应用在安装和运行期间连接、下载和上传数据到互联网',
		notifications: '通知',
		notifications_label: '发送通知的权限',
		notifications_desc:
			'允许此应用发送通知，包括 Olares 消息、短信、电子邮件和第三方即时通讯消息。您可以在设置中配置这些偏好。',
		analytics: '分析',
		analytics_label: '此应用使用内置工具收集您需要的网站分析指标',
		analytics_desc:
			'不收集任何个人信息，所有网页数据均为匿名并存储在 Olares 本地',
		websocket: 'Websocket',
		websocket_label: '允许使用 WebSocket 进行浏览器和 Olares 之间双向交互通信',
		websocket_desc: '允许此应用通过 WebSocket 向您浏览器中的网页发送内容',
		secret: '密钥',
		secret_label: '使用 Vault 创建和管理应用配置和密钥',
		secret_desc:
			'Vault 是 Olares 的密钥管理应用，用于存储、管理和同步敏感信息，如 API 密钥、数据库凭证和环境变量等',
		knowledgebase: '知识库',
		knowledgebase_label: '访问本地知识库',
		knowledgebase_desc: '允许此应用使用您在 Olares 本地知识库中存储的个人数据',
		search_label: '此应用支持 Olares 的全文搜索引擎',
		relational_database: '关系数据库',
		relational_database_label: '此应用程序使用 PostgreSQL 作为其关系数据库。',
		relational_database_desc: 'PostgreSQL 由 Olares 中间件服务提供。',
		document_database: '文档数据库',
		document_database_label: '此应用使用 MongoDB 作为文档数据库。',
		document_database_desc: 'MongoDB 由 Olares 中间件服务提供。',
		key_value_database: '键值数据库',
		key_value_database_label: '此应用使用 Redis 作为其键值数据库。',
		key_value_database_desc: 'Redis 由 Olares 中间件服务提供。',
		cluster_app_label: '同一 Olares 集群中的所有用户可以共享该应用',
		entrance: '入口',
		entrance_visibility_label:
			'此应用不同可见性入口的数量：{desktopSize} 个可见，{backendSize} 个不可见',
		entrance_visibility_desc_first: '可见入口允许您通过桌面访问应用的页面',
		entrance_visibility_desc_second:
			'不可见入口通常在后台运行，用于应用与其他应用交互',
		entrance_auth_level_label:
			'不同认证级别的入口数量：{publicSize} 个公开，{privateSize} 个私有',
		entrance_auth_level_desc:
			'公开入口可供互联网上的任何人访问。私有入口需要激活 Tailscale 才能访问。',
		entrance_two_factor_label: '这些入口需要双因素认证才能访问：{twoFactor}'
	},
	error: {
		index: '错误',
		network_error: '网络错误。请稍后重试。',
		unknown_error: '未知错误',
		failed_get_user_role: '获取用户角色失败',
		only_be_installed_by_the_admin: '此应用程序只能由管理员安装',
		not_admin_role_install_middleware:
			'此应用为中间件组件。请联系您的 Olares 管理员安装。',
		not_admin_role_install_cluster_app:
			'此应用为共享应用。请联系您的 Olares 管理员安装后才能使用。',
		failed_to_get_os_version: '获取 Olares 版本失败',
		app_is_not_compatible_terminus_os: '与您的 Olares 不兼容',
		failed_to_get_user_resource: '获取用户资源失败',
		user_not_enough_cpu: '配额 CPU 不足',
		user_not_enough_memory: '配额内存不足',
		failed_to_get_system_resource: '获取系统资源失败',
		need_to_install_dependent_app_first: '需要先安装依赖应用',
		terminus_not_enough_cpu: 'Olares 集群 CPU 不足',
		terminus_not_enough_memory: 'Olares 集群内存不足',
		terminus_not_enough_disk: 'Olares 集群磁盘空间不足',
		terminus_not_enough_gpu: 'Olares 集群 GPU 不足',
		app_info_get_failure: '应用信息获取失败',
		operation_preform_failure: '操作失败',
		cluster_not_support_platform: '此应用/推荐/大模型不支持当前 Olares 平台',
		app_install_conflict:
			'此应用与您已安装的应用冲突。请先卸载以下应用，然后再继续安装。',

		app_not_install_details: '应用名称：{name}，版本：{version}',
		middleware_not_install_details: '中间件：{name}，版本：{version}',
		app_install_conflict_details: '{name}',
		failed_to_get_your_olares_platform_architecture: '获取系统的架构失败',
		no_supportarch_specified_in_app_chart: '未找到应用支持的系统架构',
		source_conflict:
			'来源冲突: 已安装了来自其他来源的同名应用。请先卸载现有应用，再从当前来源进行安装。',
		the_dependent_app_must_be_installed_first:
			'需要先安装依赖应用。请联系您的 Olares 管理员安装所需的共享应用，然后再尝试安装。'
	},
	my: {
		all: '全部',
		olares: 'Olares',
		local: '本地',
		unable_to_install_app: '无法安装应用',
		update_all: '全部更新',
		available_updates: '可用更新',
		everything_up_to_date: '恭喜！所有应用已是最新版本。',
		no_upload_chart_tips: '您尚未上传任何自定义 chart 文件',
		no_installed_app_tips: '您尚未安装任何内容',
		no_logs: '无可用的安装日志',
		sure_to_uninstall_the_app: '您确定要卸载“{title}”吗？',
		upload_custom_chart: '上传自定义 chart',
		logs: '日志'
	},
	manage: '管理',
	updates: '更新',
	search_result_for: '"{content}"的搜索结果',
	//temp
	discover_amazing_apps: '发现精彩应用',
	featured_app: '精选应用',
	control_your_own_social_network: '掌控自己的社交网络。',
	decentralized_social_media: '去中心化社交媒体',
	get_started: '开始',
	booster_your_software_development_productivity: '提升软件开发效率。',
	enjoy_coding_now: '即刻享受编程乐趣',
	diving_into_ai_image_generation: '探索 AI 图像生成。',
	unleashing_your_creativity: '释放你的创造力！',
	mastering_your_photo_with_a_personal_library: '打造个人图库，掌控您的照片',
	organize_your_memories: '整理您的回忆',
	nocode_solution_for_your_data_management: '无代码数据管理方案。',
	build_databases_as_spreadsheets: '轻松构建数据库',

	top_app_in: '{category}最受欢迎应用',
	latest_app_in: '{category}最新应用',
	recommend_app_in: '{category}推荐应用',
	app_list: '应用列表',
	top_app_on_terminus: 'Olares 最受欢迎应用',
	latest_app_on_terminus: 'Olares 最新应用',
	'Are you sure you want to delete this app chart from Local Sources?':
		'你确定要从本地源中删除这个应用的安装包吗？',
	'Quickly find the apps you need': '快速查找你想要的应用',
	'Enter a keyword to start searching': '输入关键词开始搜索',
	no_find_app_in_search:
		'未找到与“{keyword}”相关的应用\n\n你可以尝试：\n- 检查关键词拼写。\n- 尝试其他关键词。\n- 通过类别浏览。',
	'Upload Chart': '上传安装包',
	'Uploading chart': '正在上传安装包“{chart}”...',
	'Upload failed for chart': '安装包“{chart}”上传失败',
	'Processing app data': '正在处理应用数据...',
	'App was successfully added to your local source. Do you want to install it now?':
		'应用“{application}”已添加到本地源。是否现在安装？',
	'An app named already exists in the local source. Do you want to update it with the one you just uploaded?':
		'本地源中已存在名为{application}的应用。是否用刚上传的安装包更新它？',
	'Incoming version:': '待传入版本：',
	'Installed version:': '已安装版本:',
	'Install later': '稍后安装',
	'Install now': '立即安装',
	'Unable to add the app. An app named with a newer version already exists in the Local Source.':
		'无法添加 “{application}”。本地源中已存在更新的版本。',
	'Adding Chart to Local Source': '正在将安装包“{chart}”添加到本地源',
	'Failed to upload chart': '安装包“{chart}”上传失败',
	'Add app success': '添加应用成功',
	'Download Raw Log': '下载日志',
	'Getting your app ready': '正在准备应用',
	'This can sometimes take a few moments. Thanks for your patience.':
		'这个过程可能需要一些时间。感谢您的耐心等待。',
	'Entrance paused': '入口已暂停',
	'Entrance to this application is currently paused. Please try resuming the app to re-start it.':
		'此应用入口当前已暂停运行。请尝试重新启动。',
	Settings: '设置',
	'Market Source': '市场来源',
	'Choose a remote market source to retrieve application information.':
		'选择一个远端市场来源用于获取应用程序内容。',
	'Source ID': '源 ID',
	Description: '描述',
	'Add Source': '添加源',
	'Source Title': '源标题',
	'Source URL': '源URL',
	'Change Source': '更换源',
	'Are you sure you want to change the market sources? This might take some time to update.':
		'您确定要更改市场来源吗？更新可能需要一些时间。',
	'Delete Source': '删除来源',
	'Are you sure you want to delete this market source? This operation is permanent and cannot be undone.':
		'您确定要删除此市场来源吗？此操作是永久性的，无法撤消。',
	'Only alphanumeric characters are allowed.': '标题仅限字母和数字',
	'Source Title should be less than 10 characters': '标题不应超过10个字符',
	'Source ID already exists. Please use a different name.':
		"数据源 ID '{sourceId}' 已存在，请使用其他名称。",
	'Source URL already exists. Please use a different source URL':
		'数据源 URL 已存在，请使用其他数据源 URL',
	'Also uninstall the shared server (affects all users)':
		'同时卸载共享服务（影响所有用户）',
	'Warning! Uninstalling the shared server will:':
		'警告！卸载共享服务器将：\n• 立即终止所有用户的应用程序访问权限\n• 永久删除所有存储的用户数据\n\n此操作无法撤销。请确认您了解这些后果，并希望继续完全卸载{appName}。',
	'Cannot delete this source. Please uninstall all apps installed from this source, then try again.':
		'无法删除此源。请先删除所有从此源安装的应用，然后重试。',
	'Configure Environment Variables': '配置环境变量',
	httpsRequired: '链接必须以 https:// 开头',
	invalidUrlFormat: '无效的 URL 格式',
	upgrade_app_version: `应用版本更新：从 {currentVersion} 升级至 {targetVersion}`,
	'This value is set by a system environment variable':
		'此值由系统环境变量设置。若要修改，请前往 “设置”>“开发者”>“系统环境”，并更改 [{envName}]',
	'Clone App': '克隆应用',
	'Only lowercase letters and numbers are allowed; special characters and spaces are not permitted.':
		'仅支持小写字母和数字，不允许特殊符号或空格',
	'The length cannot exceed characters.': '长度不能超过{length}个字符',
	'Removal of the installation package is prohibited as there are cloned versions of the current application.':
		'当前应用存在克隆应用，禁止移除安装包。',
	'Access to Files': '访问文件',
	'Data, Cache and User directories': '数据、缓存和用户目录',
	'Data and Cache directories': '数据和缓存目录',
	'Data and User directories': '数据和用户目录',
	'Cache and User directories': '缓存和用户目录',
	'User directory': '用户目录',
	'Data directory': '数据目录',
	'Cache directory': '缓存目录',
	'Provide Public Entrance': '提供公共入口',
	'This app provides an publicly accessible entrance that does not require authentication.':
		'此应用提供可公开访问的入口，无需进行身份验证。',
	"All traffic to this entrance may consume your reverse proxy's bandwidth.":
		'进入该入口的所有流量可能会消耗您的反向代理带宽。',
	'No Visible Entrance': '无可见入口',
	'This is a background service with no user interface. It provides an API for other apps to interact with.':
		'这是一项无用户界面的后台服务，为其他应用提供可供交互的API接口。',
	'Shared App': '共享应用',
	'This app is shared by all users in the same Olares cluster.':
		'此应用由同一Olares集群中的所有用户共享。',
	'Administrator Only': '仅管理员可用',
	'This app requires Olares administrator privileges to install.':
		'安装此应用需要拥有Olares管理员权限。',
	'Multiple Instances': '多实例',
	'This app can be cloned into multiple independent instances after installation.':
		'此应用安装后可被克隆为多个独立实例。',
	'Using System Environment Variables': '使用系统环境变量',
	'This app retrieves its env values from the following system environment variables.':
		'此应用从下述系统环境变量中获取其环境变量值。',
	'Connect to Other Apps': '连接至其他应用',
	'This app requires permission to call following providers to extend its functionality':
		'此应用需要调用下述服务提供者的权限，以扩展其功能。',
	'Using Middleware on Olares': '在Olares上使用中间件',
	'This app requires the following middleware.': '此应用需要下述中间件。',
	'Expose Ports for Remote Access': '暴露端口以供远程访问',
	'This app opens the following ports for remote access when VPN is enabled':
		'启用VPN时，此应用会开放下述端口以供远程访问。',
	'New Application Title': '新的应用名称',
	'Please provide a unique title for your cloned application to avoid confusion with existing ones.':
		'请为克隆的应用设定一个独特的名称，以免与已有的应用混淆。',
	'Please provide a unique title for the app entrance on Desktop':
		'请为应用的桌面入口设定一个独特的名称',
	'Desktop shortcut name for': '{name}的桌面图标名称',
	Clone: '克隆',
	'Also stop the shared server （affects all users)':
		'同时停止共享服务器（影响所有用户）',
	'Are you sure you want to stop the shared app':
		'你确定要停止该共享应用"{app}"吗？',
	'Are you sure you want to resume the shared app "{app}" ?（affects all users）':
		'你确定要恢复该共享应用"{app}"吗？（影响所有用户）'
};
