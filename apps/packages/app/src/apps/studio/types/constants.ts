export const cpuOptions = ['50m', '100m', '200m', '500m', '1', '2', '4'];
export const cpuUnitOptions = ['m', 'core'];
export const maxCpu = 4 * 1000; // in millicores
export const memoryOptions = ['256Mi', '512Mi', '1Gi', '2Gi', '4Gi', '8Gi'];
export const memoryUnitOptions = ['Mi', 'Gi'];
export const maxMemory = 8 * 1024; // in Mi
export const envOptions = [
	{
		label: 'beclab/go-dev-1.21:1.0.0',
		value: 'beclab/go-dev-1.21:1.0.0'
	},
	{
		label: 'beclab/node20-ts-dev:1.0.0',
		value: 'beclab/node20-ts-dev:1.0.0'
	},
	{
		label: 'beclab/python-dev-3.14:1.0.0',
		value: 'beclab/python-dev-3.14:1.0.0'
	},
	{
		label: 'beclab/cuda12.8-python-3.14:1.0.0',
		value: 'beclab/cuda12.8-python-3.14:1.0.0'
	},
	{
		label: 'beclab/cuda12.8:1.0.0',
		value: 'beclab/cuda12.8:1.0.0'
	}
];

export const diskOptions = ['Gi', 'Mi'];

// export const FilesOption: Record<
//   OPERATE_ACTION,
//   {
//     name: string;
//     icon: string;
//   }
// > = {
//   [OPERATE_ACTION.ADD_FOLDER]: {
//     name: 'Add Folder',
//     icon: 'sym_r_create_new_folder',
//   },
//   [OPERATE_ACTION.ADD_FILE]: {
//     name: 'Add File',
//     icon: 'sym_r_note_add',
//   },
//   [OPERATE_ACTION.RENAME]: {
//     name: 'Rename',
//     icon: 'sym_r_edit_square',
//   },
//   [OPERATE_ACTION.DELETE]: {
//     name: 'Delete',
//     icon: 'sym_r_delete',
//   },
// };
