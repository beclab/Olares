import { boot } from 'quasar/wrappers';
import { install as VueMonacoEditorPlugin } from '@guolao/vue-monaco-editor';
import * as monaco from 'monaco-editor';
import { loader } from '@guolao/vue-monaco-editor';
import { configureMonacoYaml } from 'monaco-yaml';

export default boot(async ({ app }) => {
	(window as any).MonacoEnvironment = {
		getWorker(_: string, label: string) {
			const getWorkerUrl = (file: string) => {
				return new URL(file, import.meta.url).href;
			};

			if (label === 'yaml') {
				return new Worker(new URL('monaco-yaml/yaml.worker', import.meta.url), {
					type: 'module'
				});
			}

			return new Worker(
				new URL('monaco-editor/esm/vs/editor/editor.worker', import.meta.url),
				{
					type: 'module'
				}
			);
		}
	};

	app.use(VueMonacoEditorPlugin, {});

	loader.config({ monaco });

	monaco.editor.defineTheme('olares-light', {
		base: 'vs',
		inherit: true,
		rules: [
			{ token: '', foreground: '000000' },
			{ token: 'comment', foreground: '8c8c8c', fontStyle: 'italic' },
			{ token: 'keyword', foreground: '000080', fontStyle: 'bold' },
			{ token: 'keyword.control', foreground: '000080', fontStyle: 'bold' },
			{ token: 'keyword.operator', foreground: '000000' },
			{ token: 'variable', foreground: '000000' },
			{ token: 'constant', foreground: '067d17' },
			{ token: 'constant.language', foreground: '000080', fontStyle: 'bold' },
			{ token: 'constant.numeric', foreground: '1750eb' },
			{ token: 'string', foreground: '067d17' },
			{ token: 'string.key', foreground: '871094' },
			{ token: 'entity.name.function', foreground: '000000' },
			{ token: 'entity.name.type', foreground: '000000' },
			{ token: 'entity.name.tag', foreground: '000080' },
			{ token: 'entity.other.attribute-name', foreground: '0000ff' },
			{ token: 'support.function', foreground: '000000' },
			{ token: 'support.constant', foreground: '871094' },
			{ token: 'support.type', foreground: '000000' },
			{ token: 'support.class', foreground: '000000' },
			{ token: 'storage', foreground: '000080', fontStyle: 'bold' },
			{ token: 'storage.type', foreground: '000080', fontStyle: 'bold' },
			{ token: 'number', foreground: '1750eb' },
			{ token: 'tag', foreground: '000080' },
			{ token: 'attribute.name', foreground: '0000ff' },
			{ token: 'attribute.value', foreground: '067d17' },
			{ token: 'type', foreground: '000000' }
		],
		colors: {
			'editor.background': '#ffffff',
			'editor.foreground': '#000000',
			'editorLineNumber.foreground': '#adadad',
			'editorLineNumber.activeForeground': '#000000',
			'editorIndentGuide.background': '#e8e8e8',
			'editorIndentGuide.activeBackground': '#b0b0b0',
			'editor.selectionBackground': '#cee7ff',
			'editor.lineHighlightBackground': '#f5f5f5',
			'editorCursor.foreground': '#000000',
			'editorWhitespace.foreground': '#e8e8e8'
		}
	});

	monaco.editor.defineTheme('olares-dark', {
		base: 'vs-dark',
		inherit: true,
		rules: [
			{ token: '', foreground: 'f8f8f2' },
			{ token: 'comment', foreground: '6272a4', fontStyle: 'italic' },
			{ token: 'keyword', foreground: 'ff79c6' },
			{ token: 'keyword.control', foreground: 'ff79c6' },
			{ token: 'keyword.operator', foreground: 'ff79c6' },
			{ token: 'variable', foreground: 'f8f8f2' },
			{ token: 'constant', foreground: 'bd93f9' },
			{ token: 'constant.numeric', foreground: 'bd93f9' },
			{ token: 'string', foreground: 'f1fa8c' },
			{ token: 'entity.name.function', foreground: '50fa7b' },
			{ token: 'entity.name.type', foreground: '8be9fd' },
			{ token: 'entity.name.tag', foreground: 'ff79c6' },
			{ token: 'entity.other.attribute-name', foreground: '50fa7b' },
			{ token: 'support.function', foreground: '8be9fd' },
			{ token: 'support.constant', foreground: 'bd93f9' },
			{ token: 'support.type', foreground: '8be9fd' },
			{ token: 'support.class', foreground: '8be9fd' },
			{ token: 'storage', foreground: 'ff79c6' },
			{ token: 'storage.type', foreground: '8be9fd' },
			{ token: 'number', foreground: 'bd93f9' },
			{ token: 'tag', foreground: 'ff79c6' },
			{ token: 'attribute.name', foreground: '50fa7b' },
			{ token: 'attribute.value', foreground: 'f1fa8c' }
		],
		colors: {
			'editor.background': '#282a36',
			'editor.foreground': '#f8f8f2',
			'editorLineNumber.foreground': '#6272a4',
			'editorLineNumber.activeForeground': '#f8f8f2',
			'editorIndentGuide.background': '#44475a',
			'editorIndentGuide.activeBackground': '#6272a4',
			'editor.selectionBackground': '#44475a',
			'editor.lineHighlightBackground': '#44475a',
			'editorCursor.foreground': '#f8f8f0',
			'editorWhitespace.foreground': '#44475a'
		}
	});

	configureMonacoYaml(monaco, {
		enableSchemaRequest: true,
		hover: true,
		completion: true,
		validate: true,
		format: true,
		schemas: [
			{
				uri: 'https://json.schemastore.org/chart.json',
				fileMatch: ['**/Chart.yaml', '**/chart.yaml']
			},
			{
				uri: 'https://json.schemastore.org/kustomization.json',
				fileMatch: ['**/kustomization.yaml', '**/kustomization.yml']
			},
			{
				uri: 'https://json.schemastore.org/github-workflow.json',
				fileMatch: ['**/.github/workflows/*.yaml', '**/.github/workflows/*.yml']
			},
			{
				uri: 'https://raw.githubusercontent.com/compose-spec/compose-spec/master/schema/compose-spec.json',
				fileMatch: [
					'**/docker-compose.yaml',
					'**/docker-compose.yml',
					'**/compose.yaml',
					'**/compose.yml'
				]
			},
			{
				uri: 'olares://schemas/olares-manifest',
				fileMatch: ['**/OlaresManifest.yaml', '**/OlaresManifest.yml'],
				schema: {
					type: 'object',
					properties: {
						metadata: {
							type: 'object',
							properties: {
								name: { type: 'string', description: 'Application name' },
								title: { type: 'string', description: 'Application title' },
								version: {
									type: 'string',
									pattern: '^\\d+\\.\\d+\\.\\d+',
									description: 'Version in SemVer format'
								},
								description: {
									type: 'string',
									description: 'Short description'
								},
								icon: { type: 'string', description: 'Application icon URL' },
								categories: {
									type: 'array',
									items: { type: 'string' },
									description: 'Application categories'
								}
							},
							required: ['name', 'title', 'version']
						},
						spec: {
							type: 'object',
							properties: {
								entrances: {
									type: 'array',
									items: {
										type: 'object',
										properties: {
											name: { type: 'string' },
											title: { type: 'string' },
											port: { type: 'integer', minimum: 1, maximum: 65535 },
											host: { type: 'string' },
											icon: { type: 'string' }
										},
										required: ['name', 'port']
									}
								},
								requiredMemory: {
									type: 'string',
									pattern: '^\\d+(Mi|Gi)$',
									description: 'Required memory (e.g., 256Mi, 1Gi)'
								},
								limitedMemory: {
									type: 'string',
									pattern: '^\\d+(Mi|Gi)$',
									description: 'Memory limit'
								},
								requiredCpu: {
									type: 'string',
									pattern: '^\\d+m?$',
									description: 'Required CPU (e.g., 100m, 1)'
								},
								limitedCpu: {
									type: 'string',
									pattern: '^\\d+m?$',
									description: 'CPU limit'
								}
							}
						}
					},
					required: ['metadata', 'spec']
				}
			}
		]
	});

	monaco.languages.registerCompletionItemProvider('yaml', {
		provideCompletionItems: (model, position) => {
			const textUntilPosition = model.getValueInRange({
				startLineNumber: position.lineNumber,
				startColumn: 1,
				endLineNumber: position.lineNumber,
				endColumn: position.column
			});

			if (
				textUntilPosition.trim().length > 0 &&
				!textUntilPosition.endsWith(' ')
			) {
				return { suggestions: [] };
			}

			const word = model.getWordUntilPosition(position);
			const range = {
				startLineNumber: position.lineNumber,
				endLineNumber: position.lineNumber,
				startColumn: word.startColumn,
				endColumn: word.endColumn
			};

			const suggestions: monaco.languages.CompletionItem[] = [
				{
					label: 'k8s-deployment',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'apiVersion: apps/v1',
						'kind: Deployment',
						'metadata:',
						'  name: ${1:app-name}',
						'  labels:',
						'    app: ${1:app-name}',
						'spec:',
						'  replicas: ${2:1}',
						'  selector:',
						'    matchLabels:',
						'      app: ${1:app-name}',
						'  template:',
						'    metadata:',
						'      labels:',
						'        app: ${1:app-name}',
						'    spec:',
						'      containers:',
						'      - name: ${1:app-name}',
						'        image: ${3:image:tag}',
						'        ports:',
						'        - containerPort: ${4:8080}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Kubernetes Deployment template',
					range
				},
				{
					label: 'k8s-service',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'apiVersion: v1',
						'kind: Service',
						'metadata:',
						'  name: ${1:service-name}',
						'spec:',
						'  selector:',
						'    app: ${2:app-name}',
						'  ports:',
						'  - protocol: TCP',
						'    port: ${3:80}',
						'    targetPort: ${4:8080}',
						'  type: ${5|ClusterIP,NodePort,LoadBalancer|}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Kubernetes Service template',
					range
				},
				{
					label: 'k8s-configmap',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'apiVersion: v1',
						'kind: ConfigMap',
						'metadata:',
						'  name: ${1:configmap-name}',
						'data:',
						'  ${2:key}: ${3:value}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Kubernetes ConfigMap template',
					range
				},
				{
					label: 'k8s-secret',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'apiVersion: v1',
						'kind: Secret',
						'metadata:',
						'  name: ${1:secret-name}',
						'type: Opaque',
						'data:',
						'  ${2:key}: ${3:base64-value}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Kubernetes Secret template',
					range
				},
				{
					label: 'helm-chart',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'apiVersion: v2',
						'name: ${1:chart-name}',
						'description: ${2:A Helm chart for Kubernetes}',
						'type: application',
						'version: ${3:0.1.0}',
						'appVersion: "${4:1.0.0}"',
						'keywords:',
						'  - ${5:keyword}',
						'maintainers:',
						'  - name: ${6:maintainer-name}',
						'    email: ${7:email@example.com}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Helm Chart.yaml template',
					range
				},
				{
					label: 'olares-manifest',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'metadata:',
						'  name: ${1:app-name}',
						'  title: ${2:App Title}',
						'  version: ${3:0.1.0}',
						'  description: ${4:App description}',
						'  icon: ${5:https://example.com/icon.png}',
						'  categories:',
						'    - ${6:Utilities}',
						'',
						'spec:',
						'  entrances:',
						'    - name: ${1:app-name}',
						'      title: ${2:App Title}',
						'      port: ${7:8080}',
						'      host: ${1:app-name}',
						'      icon: ${5:https://example.com/icon.png}',
						'  ',
						'  requiredMemory: ${8:256Mi}',
						'  limitedMemory: ${9:512Mi}',
						'  requiredCpu: ${10:100m}',
						'  limitedCpu: ${11:500m}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Olares Application Manifest template',
					range
				},
				{
					label: 'docker-compose-service',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'${1:service-name}:',
						'  image: ${2:image:tag}',
						'  container_name: ${3:container-name}',
						'  ports:',
						'    - "${4:8080}:${5:80}"',
						'  environment:',
						'    - ${6:ENV_VAR}=${7:value}',
						'  volumes:',
						'    - ${8:./data}:${9:/app/data}',
						'  restart: ${10|unless-stopped,always,on-failure|}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Docker Compose service definition',
					range
				},
				{
					label: 'env-vars',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: ['- name: ${1:ENV_NAME}', '  value: "${2:value}"'].join(
						'\n'
					),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Environment variable definition',
					range
				},
				{
					label: 'volume-mount',
					kind: monaco.languages.CompletionItemKind.Snippet,
					insertText: [
						'- name: ${1:volume-name}',
						'  mountPath: ${2:/path/in/container}'
					].join('\n'),
					insertTextRules:
						monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
					documentation: 'Volume mount definition',
					range
				}
			];

			return { suggestions };
		}
	});
});
