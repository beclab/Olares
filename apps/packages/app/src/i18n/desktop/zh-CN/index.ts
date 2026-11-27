import market from '../../market/zh-CN';

export default {
	launch_input_placeholder: '搜索',
	launch_no_result: '没有结果',
	page_404: '这里还没有内容',
	landscape_message: '请将设备旋转为竖屏查看',
	go_home: '返回首页',
	system_app: {
		Files: '文件管理器',
		Market: '应用市场',
		Dashboard: '仪表盘',
		Settings: '设置',
		'Control Hub': '控制面板',
		Desktop: '桌面'
	},
	confirmation: '无法打开应用',
	message_desktop: {
		suspended: '应用已暂停。请在设置或应用市场中恢复后重试。',
		crashed: '应用已崩溃，正在重启',
		delete_app: '删除 {appName}？'
	},
	upgrade: {
		success_title: '系统已更新',
		success_message: '更新完成',
		fail_title: '无法更新',
		fail_message: '无法更新，请重试',
		running_title: '正在更新系统…',
		running_message: '可能需要几分钟'
	},
	dock_operation: {
		open: '打开',
		show: '显示',
		hide: '隐藏',
		keepInDock: '在 Dock 中保留',
		removeFromDock: '从 Dock 中移除',
		quit: '退出'
	},
	'Load more': '加载更多',
	'There is currently no news': '暂无消息',
	'Your changes may not be saved.': '未保存的更改可能会丢失',
	'Clear all': '清空所有',
	notifications: '通知',
	desktop_layout: {
		reset_layout_title: '重置桌面布局',
		reset_layout_button: '重置',
		reset_confirm_title: '确认重置桌面布局？',
		reset_confirm_warning:
			'启动台和程序坞将恢复为默认布局，自定义图标位置和程序坞项目都会被重置。此操作不可撤销。',
		reset_confirm_note: '应用数据不会被删除。',
		reset_failed: '无法重置，请重试'
	},
	desktop_settings: {
		title: '桌面设置',
		bookmarks: '书签',
		bookmarks_desc: '将常用网站添加到启动台，方便快速访问',
		add_bookmark: '添加书签',
		edit_bookmark: '编辑书签',
		delete_bookmark_title: '移除书签？',
		delete_bookmark_msg: '{name} 将从启动台和程序坞中移除',
		no_bookmarks: '暂无书签',
		loading: '加载中…',
		name: '名称',
		url: 'URL',
		icon_optional: '图标 URL（可选）',
		open_mode: '打开方式',
		open_iframe: '内嵌窗口',
		open_window: '新标签页',
		required_fields: '名称和 URL 都不能为空',
		name_exists: '名称已被应用或书签使用',
		save_failed: '无法保存，请重试',
		advanced: '高级',
		menu_open: '桌面设置…',
		iframe_blocked_title: '无法嵌入该网站',
		iframe_blocked_msg: '此网站不允许在其他页面中显示。请在新标签页中打开。',
		open_in_new_tab: '在新标签页打开'
	}
};
