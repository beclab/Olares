import { t } from 'src/boot/studio-i18n';

export const ruleConfig = {
	appNameRules: {
		rules: [
			(val) => (val && val.length > 0) || t('home_appname_rules_1'),
			(val) => /^[A-Za-z].*/.test(val) || t('home_appname_rules_2'),
			(val) =>
				/^[a-zA-Z][a-zA-Z0-9 ._-]{0,30}$/.test(val) || t('home_appname_rules_3')
		],
		placeholder: t('home_appname_rules_1')
	},
	imageName: {
		rules: [(val) => (val && val.length > 0) || t('image_rule')],
		placeholder: t('image_rule')
	},

	startCommand: {
		rules: [(val) => (val && val.length > 0) || t('start_command_rule')],
		placeholder: t('start_command_rule')
	},

	startCmdArgs: {
		rules: [(val) => (val && val.length > 0) || t('start_cmd_args_rule')],
		placeholder: t('start_cmd_args_rule')
	},

	websitePort: {
		rules: [
			(val) => (val && val.length > 0) || t('home_entrance_port_rules_1'),
			(val) => (val > 0 && val <= 65535) || t('home_entrance_port_rules_2')
		],
		placeholder: t('home_entrance_port_rules_1')
	},

	cpu: {
		rules: [(val) => (val && val.length > 0) || t('cpu_rule')],
		placeholder: t('cpu_rule')
	},
	memory: {
		rules: [(val) => (val && val.length > 0) || t('memory_rule')],
		placeholder: t('memory_rule')
	},

	envConfigKey: {
		rules: [(val) => (val && val.length > 0) || t('enter_input')],
		placeholder: t('enter_input')
	},

	envConfigValue: {
		rules: [(val) => (val && val.length > 0) || t('enter_input')],
		placeholder: t('enter_input')
	},

	hostPath: {
		rules: [
			(val) => (val && val.length > 0) || t('host_path_rule'),
			(val) => /^\/.*$/.test(val) || t('host_path_rule_2')
		],
		placeholder: t('host_path_rule')
	},

	containerPath: {
		rules: [
			(val) => (val && val.length > 0) || t('host_container_rule'),
			(val) => /^\/.+$/.test(val) || t('host_container_rule_2')
		],
		placeholder: t('host_container_rule')
	},

	file: {
		rules: [
			(val) => (val && val.length > 0) || t('file_name_rule'),
			(val) => !/[\\/:*?"<>|]/.test(val) || t('file_name_rule_2')
		],
		placeholder: t('file_name_rule')
	},

	env: {
		rules: [(val) => (val && val.length > 0) || t('image_env')],
		placeholder: t('image_env')
	},

	volume: {
		rules: [
			(val) => (val && val.length > 0) || t('image_volume'),
			(val) => /^[1-9]\d*$/.test(val) || t('image_volume_rule')
		],
		placeholder: t('image_volume')
	},
	ports: {
		rules: [
			(val) => {
				if (!val || val.length === 0) return true;
				const trimmedVal = val.replace(/\s+/g, '');
				return /^[0-9,]+$/.test(trimmedVal) || t('image_ports_rule_1');
			},
			(val) => {
				if (!val || val.length === 0) return true;
				const ports = val
					.split(',')
					.map((p) => p.trim())
					.filter((p) => p);
				const allValid = ports.every((port) => {
					const num = parseInt(port);
					return !isNaN(num) && num >= 1 && num <= 65535;
				});
				return allValid || t('image_ports_rule_2');
			},
			(val) => {
				if (!val || val.length === 0) return true;
				const ports = val
					.split(',')
					.map((p) => p.trim())
					.filter((p) => p);
				const forbiddenPorts = ['5000', '80', '443'];
				const hasForbidden = ports.some((port) =>
					forbiddenPorts.includes(port)
				);
				return !hasForbidden || t('image_ports_rule_3');
			}
		],
		placeholder: t('image_ports')
	}
};
