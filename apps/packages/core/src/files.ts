// All extensions that should be treated as plain-text (preview/edit).
// No leading dot, lowercase only. Extend this list when adding text-readable formats.
export const TEXT_FILE_EXTENSIONS: string[] = [
	// Plain text & rich text
	'txt',
	'text',
	'log',
	'rtf',

	// Markdown & documentation
	'md',
	'markdown',
	'mdx',
	'rst',
	'adoc',
	'asciidoc',
	'org',
	'tex',
	'bib',

	// Data / markup
	'json',
	'json5',
	'jsonc',
	'jsonl',
	'ndjson',
	'ipynb',
	'xml',
	'xhtml',
	'html',
	'htm',
	'svg',
	'csv',
	'tsv',

	// Subtitle
	'srt',
	'vtt',
	'ass',
	'ssa',
	'sub',

	// Configuration
	'yaml',
	'yml',
	'toml',
	'ini',
	'conf',
	'cfg',
	'config',
	'env',
	'plist',
	'properties',
	'editorconfig',
	'htaccess',
	'lock',

	// Shell / scripts
	'sh',
	'bash',
	'zsh',
	'fish',
	'ps1',
	'bat',
	'cmd',

	// VCS / container / build
	'gitignore',
	'gitattributes',
	'dockerfile',
	'dockerignore',
	'makefile',
	'mk',
	'cmake',

	// Web / frontend
	'css',
	'scss',
	'sass',
	'less',
	'styl',
	'style',
	'js',
	'mjs',
	'cjs',
	'jsx',
	'ts',
	'tsx',
	'vue',

	// JVM languages
	'java',
	'kt',
	'scala',
	'groovy',
	'gradle',
	'clj',
	'cljs',
	'cljc',

	// C / C++ & systems
	'c',
	'cpp',
	'cc',
	'cxx',
	'h',
	'hpp',
	'hh',
	'hxx',
	'm',
	'mm',
	'asm',
	's',

	// Mainstream scripting / server-side
	'py',
	'php',
	'rb',
	'pl',
	'pm',
	'lua',
	'r',
	'sql',
	'go',
	'rs',
	'swift',
	'cs',
	'vbp',
	'vbs',
	'dart',
	'sol',

	// Functional languages
	'hs',
	'ex',
	'exs',
	'erl',
	'ml',
	'mli',
	'fs',
	'fsi',
	'fsx',

	// Other common
	'for',
	'f90',
	'f95',
	'pas',
	'dpr',
	'nim',
	'zig',
	'sb3'
];

const FILE_ICON_CONFIG: { [key: string]: string[] } = {
	image: [
		'png',
		'jpg',
		'jpeg',
		'bmp',
		'gif',
		'heic',
		'webp',
		'svg',
		'tif',
		'tiff',
		'raw'
	],
	txt: TEXT_FILE_EXTENSIONS,
	epub: ['epub'],
	excel: ['xls', 'xlsx'],
	word: ['doc', 'docx'],
	pdf: ['pdf'],
	ppt: ['ppt', 'pptx'],
	zip: ['rar', 'zip', '7z', 'tar', 'gz', 'bz2'],
	video: [
		'mp4',
		'm2v',
		'mkv',
		'rmvb',
		'wmv',
		'avi',
		'flv',
		'mov',
		'm4v',
		'ape',
		'webm',
		'vob',
		'mpg',
		'3gp',
		'rm'
	],
	audio: ['mp3', 'wav', 'm4a', 'flac', 'ogg', 'aac']
};

export function getFileIcon(fileName: string): string {
	if (!fileName) return 'other';

	const extension = fileName.split('.').pop()?.toLowerCase();
	if (!extension) return 'other';

	for (const iconType in FILE_ICON_CONFIG) {
		const extensions = FILE_ICON_CONFIG[iconType];
		if (extensions.indexOf(extension) !== -1) {
			return iconType;
		}
	}

	return 'other';
}

// Type labels follow the convention used by Windows Explorer's "Type" column:
//   - images use `XXX image`
//   - videos use `XXX Video` / `Video Clip`
//   - audio  use `XXX Audio File` / `Wave Sound`
//   - archives use Windows-style names (e.g. `Compressed (zipped) Folder`, `WinRAR archive`)
//   - source code / scripts use `XXX Source File` / `XXX Script File`
//   - configuration uses `Configuration Settings` / `XXX File`
const FILE_TYPE_CONFIG: { [key: string]: string[] } = {
	// Office & page-layout documents
	'Microsoft Word 97-2003 Document': ['doc'],
	'Microsoft Word Document': ['docx'],
	'OpenDocument Text': ['odt'],
	'Pages Document': ['pages'],
	'Rich Text Document': ['rtf'],
	'PDF Document': ['pdf'],
	'Microsoft PowerPoint Presentation': ['ppt', 'pptx'],
	'Microsoft Excel Worksheet': ['xls', 'xlsx'],
	'Microsoft Excel Comma Separated Values File': ['csv'],
	'Tab Separated Values File': ['tsv'],

	// Markup & web
	'XML Document': ['xml'],
	'HTML Document': ['xhtml', 'html', 'htm'],

	// Images
	'JPEG image': ['jpg', 'jpeg', 'jpe', 'jfif', 'jfif-tbnl'],
	'PNG image': ['png'],
	'GIF image': ['gif'],
	'RAW image': ['raw'],
	'Bitmap image': ['bmp'],
	'SVG Document': ['svg'],
	'WebP image': ['webp'],
	'HEIC image': ['heic'],
	'TIFF image': ['tif', 'tiff'],

	// E-books & plain text
	'EPUB File': ['epub'],
	'Text Document': ['txt', 'text'],
	'Markdown File': ['md', 'markdown'],
	'MDX Document': ['mdx'],
	'JSON File': ['json', 'json5', 'jsonc', 'jsonl', 'ndjson'],
	'Jupyter Notebook': ['ipynb'],
	'reStructuredText Document': ['rst'],
	'AsciiDoc Document': ['adoc', 'asciidoc'],
	'Org Mode Document': ['org'],
	'LaTeX Document': ['tex'],
	'BibTeX Bibliography': ['bib'],

	// Subtitles
	'Subtitle File': ['srt', 'vtt', 'ass', 'ssa', 'sub'],

	// Archives (Windows Explorer naming)
	'Compressed (zipped) Folder': ['zip'],
	'WinRAR archive': ['rar'],
	'7-Zip archive': ['7z'],
	'TAR archive': ['tar'],
	'GZIP archive': ['gz'],
	'BZIP2 archive': ['bz2'],

	// Audio
	'MP3 Audio File': ['mp3'],
	'M4A Audio File': ['m4a'],
	'Wave Sound': ['wav'],
	'AAC Audio File': ['aac'],
	'FLAC Audio File': ['flac'],
	'OGG Audio File': ['ogg'],
	"Monkey's Audio File": ['ape'],

	// Video
	'MP4 Video': ['mp4', 'm4v'],
	'Video Clip': [
		'avi',
		'mov',
		'wmv',
		'm2v',
		'mkv',
		'rmvb',
		'flv',
		'webm',
		'vob',
		'mpg',
		'3gp',
		'rm'
	],

	// Source code
	'Python Source File': ['py'],
	'C Source File': ['c'],
	'C++ Source File': ['cpp', 'cc', 'cxx'],
	'C/C++ Header File': ['h', 'hpp', 'hh', 'hxx'],
	'Java Source File': ['java'],
	'JavaScript File': ['js', 'mjs', 'cjs'],
	'JavaScript JSX File': ['jsx'],
	'TypeScript File': ['ts'],
	'TypeScript JSX File': ['tsx'],
	'Vue Component': ['vue'],
	'PHP Source File': ['php'],
	'SQL File': ['sql'],
	'Cascading Style Sheet Document': [
		'css',
		'scss',
		'sass',
		'less',
		'styl',
		'style'
	],
	'Shell Script File': ['sh', 'bash', 'zsh', 'fish'],
	'VBScript File': ['vbs'],
	'Swift Source File': ['swift'],
	'Objective-C Source File': ['m', 'mm'],
	'C# Source File': ['cs'],
	'F# Source File': ['fs', 'fsi', 'fsx'],
	'Go Source File': ['go'],
	'Ruby Source File': ['rb'],
	'Perl Script File': ['pl'],
	'Perl Module File': ['pm'],
	'Lua Source File': ['lua'],
	'Visual Basic Project File': ['vbp'],
	'Fortran Source File': ['for', 'f90', 'f95'],
	'Pascal Source File': ['pas'],
	'Delphi Project File': ['dpr'],
	'Rust Source File': ['rs'],
	'Assembly Source File': ['asm', 's'],
	'Kotlin Source File': ['kt'],
	'R Source File': ['r'],
	'Scratch Project File': ['sb3'],
	'Dart Source File': ['dart'],
	'Scala Source File': ['scala'],
	'Groovy Source File': ['groovy'],
	'Gradle Build File': ['gradle'],
	'Clojure Source File': ['clj', 'cljs', 'cljc'],
	'Haskell Source File': ['hs'],
	'Elixir Source File': ['ex', 'exs'],
	'Erlang Source File': ['erl'],
	'OCaml Source File': ['ml', 'mli'],
	'Nim Source File': ['nim'],
	'Zig Source File': ['zig'],
	'Solidity Source File': ['sol'],

	// Configuration & scripts (.bat / .cmd are Windows batch files)
	'YAML File': ['yaml', 'yml'],
	'TOML File': ['toml'],
	'Configuration Settings': ['config', 'ini', 'conf', 'cfg'],
	'Environment File': ['env'],
	'Property List File': ['plist'],
	'Log File': ['log'],
	'Properties File': ['properties'],
	'EditorConfig File': ['editorconfig'],
	'Apache Configuration File': ['htaccess'],
	'Lock File': ['lock'],
	'Windows PowerShell Script': ['ps1'],
	'Windows Batch File': ['bat', 'cmd'],
	'Git Ignore File': ['gitignore'],
	'Git Attributes File': ['gitattributes'],
	Dockerfile: ['dockerfile'],
	'Docker Ignore File': ['dockerignore'],
	Makefile: ['makefile', 'mk'],
	'CMake File': ['cmake'],

	// Installers / executables
	'Android Package': ['apk'],
	'Android Extended Package': ['xapk'],
	'Android App Bundle': ['aab'],
	'iOS App Package': ['ipa'],
	'Windows Executable': ['exe'],
	'Windows Installer Package': ['msi'],
	'Windows Application Package': ['appx', 'msix'],
	'macOS Disk Image': ['dmg'],
	'macOS Installer Package': ['pkg'],
	'macOS Application': ['app'],
	'Debian Package': ['deb'],
	'Red Hat Package': ['rpm'],
	'Snap Package': ['snap'],
	'Flatpak Package': ['flatpak'],
	'AppImage Application': ['appimage']
};

export function getFileType(fileName: string): string {
	if (!fileName) return 'blob';

	const extension = fileName.split('.').pop()?.toLowerCase();
	if (!extension) return 'blob';

	for (const fileType in FILE_TYPE_CONFIG) {
		const extensions = FILE_TYPE_CONFIG[fileType];
		if (extensions.indexOf(extension) !== -1) {
			return fileType;
		}
	}

	return 'blob';
}

// Maps the Windows-style English labels in FILE_TYPE_CONFIG to stable snake_case
// i18n identifiers. Usage: i18n.global.t(getFileTypeI18nKey(fileName))
const FILE_TYPE_I18N_ID: { [key: string]: string } = {
	// Office & page-layout documents
	'Microsoft Word 97-2003 Document': 'word_97_2003',
	'Microsoft Word Document': 'word',
	'OpenDocument Text': 'odt',
	'Pages Document': 'pages',
	'Rich Text Document': 'rtf',
	'PDF Document': 'pdf',
	'Microsoft PowerPoint Presentation': 'powerpoint',
	'Microsoft Excel Worksheet': 'excel',
	'Microsoft Excel Comma Separated Values File': 'csv',
	'Tab Separated Values File': 'tsv',

	// Markup & web
	'XML Document': 'xml',
	'HTML Document': 'html',

	// Images
	'JPEG image': 'jpeg',
	'PNG image': 'png',
	'GIF image': 'gif',
	'RAW image': 'raw',
	'Bitmap image': 'bmp',
	'SVG Document': 'svg',
	'WebP image': 'webp',
	'HEIC image': 'heic',
	'TIFF image': 'tiff',

	// E-books & plain text
	'EPUB File': 'epub',
	'Text Document': 'txt',
	'Markdown File': 'md',
	'MDX Document': 'mdx',
	'JSON File': 'json',
	'Jupyter Notebook': 'ipynb',
	'reStructuredText Document': 'rst',
	'AsciiDoc Document': 'adoc',
	'Org Mode Document': 'org',
	'LaTeX Document': 'latex',
	'BibTeX Bibliography': 'bibtex',

	// Subtitles
	'Subtitle File': 'subtitle',

	// Archives
	'Compressed (zipped) Folder': 'zip',
	'WinRAR archive': 'rar',
	'7-Zip archive': 'seven_zip',
	'TAR archive': 'tar',
	'GZIP archive': 'gzip',
	'BZIP2 archive': 'bzip2',

	// Audio
	'MP3 Audio File': 'mp3',
	'M4A Audio File': 'm4a',
	'Wave Sound': 'wav',
	'AAC Audio File': 'aac',
	'FLAC Audio File': 'flac',
	'OGG Audio File': 'ogg',
	"Monkey's Audio File": 'ape',

	// Video
	'MP4 Video': 'mp4',
	'Video Clip': 'video',

	// Source code
	'Python Source File': 'python',
	'C Source File': 'c',
	'C++ Source File': 'cpp',
	'C/C++ Header File': 'c_header',
	'Java Source File': 'java',
	'JavaScript File': 'js',
	'JavaScript JSX File': 'jsx',
	'TypeScript File': 'ts',
	'TypeScript JSX File': 'tsx',
	'Vue Component': 'vue',
	'PHP Source File': 'php',
	'SQL File': 'sql',
	'Cascading Style Sheet Document': 'css',
	'Shell Script File': 'sh',
	'VBScript File': 'vbs',
	'Swift Source File': 'swift',
	'Objective-C Source File': 'objective_c',
	'C# Source File': 'csharp',
	'F# Source File': 'fsharp',
	'Go Source File': 'go',
	'Ruby Source File': 'ruby',
	'Perl Script File': 'perl',
	'Perl Module File': 'perl_module',
	'Lua Source File': 'lua',
	'Visual Basic Project File': 'vbp',
	'Fortran Source File': 'fortran',
	'Pascal Source File': 'pascal',
	'Delphi Project File': 'delphi',
	'Rust Source File': 'rust',
	'Assembly Source File': 'asm',
	'Kotlin Source File': 'kotlin',
	'R Source File': 'r',
	'Scratch Project File': 'scratch',
	'Dart Source File': 'dart',
	'Scala Source File': 'scala',
	'Groovy Source File': 'groovy',
	'Gradle Build File': 'gradle',
	'Clojure Source File': 'clojure',
	'Haskell Source File': 'haskell',
	'Elixir Source File': 'elixir',
	'Erlang Source File': 'erlang',
	'OCaml Source File': 'ocaml',
	'Nim Source File': 'nim',
	'Zig Source File': 'zig',
	'Solidity Source File': 'solidity',

	// Configuration & scripts
	'YAML File': 'yaml',
	'TOML File': 'toml',
	'Configuration Settings': 'config',
	'Environment File': 'env',
	'Property List File': 'plist',
	'Log File': 'log',
	'Properties File': 'properties',
	'EditorConfig File': 'editorconfig',
	'Apache Configuration File': 'htaccess',
	'Lock File': 'lock',
	'Windows PowerShell Script': 'powershell',
	'Windows Batch File': 'bat',
	'Git Ignore File': 'gitignore',
	'Git Attributes File': 'gitattributes',
	Dockerfile: 'dockerfile',
	'Docker Ignore File': 'dockerignore',
	Makefile: 'makefile',
	'CMake File': 'cmake',

	// Installers / executables
	'Android Package': 'apk',
	'Android Extended Package': 'xapk',
	'Android App Bundle': 'aab',
	'iOS App Package': 'ipa',
	'Windows Executable': 'exe',
	'Windows Installer Package': 'msi',
	'Windows Application Package': 'appx',
	'macOS Disk Image': 'dmg',
	'macOS Installer Package': 'pkg',
	'macOS Application': 'macos_app',
	'Debian Package': 'deb',
	'Red Hat Package': 'rpm',
	'Snap Package': 'snap',
	'Flatpak Package': 'flatpak',
	'AppImage Application': 'appimage'
};

const FILE_TYPE_I18N_PREFIX = 'files.file_types.';

/**
 * Returns the i18n key path for the file-type label of the given file name,
 * e.g. `files.file_types.pdf`. The caller can pass it directly to
 * `i18n.global.t()`. Returns `files.file_types.unknown` for unrecognized types.
 */
export function getFileTypeI18nKey(fileName: string): string {
	const label = getFileType(fileName);
	const id = FILE_TYPE_I18N_ID[label];
	return FILE_TYPE_I18N_PREFIX + (id ?? 'unknown');
}

export const txtReadonlyTypes = ['log'];
