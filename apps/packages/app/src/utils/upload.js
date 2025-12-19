export function scanFiles(dt) {
	return new Promise((resolve) => {
		let reading = 0;
		const contents = [];

		if (dt.items !== undefined) {
			for (let item of dt.items) {
				if (
					item.kind === 'file' &&
					typeof item.webkitGetAsEntry === 'function'
				) {
					const entry = item.webkitGetAsEntry();
					readEntry(entry);
				}
			}
		} else {
			resolve(dt.files);
		}

		function readEntry(entry, directory = '') {
			if (entry.isFile) {
				reading++;
				entry.file((file) => {
					reading--;

					file.fullPath = `${directory}${file.name}`;
					contents.push(file);

					if (reading === 0) {
						resolve(contents);
					}
				});
			} else if (entry.isDirectory) {
				const dir = {
					isDir: true,
					size: 0,
					fullPath: `${directory}${entry.name}`,
					name: entry.name
				};

				contents.push(dir);

				readReaderContent(entry.createReader(), `${directory}${entry.name}`);
			}
		}

		function readReaderContent(reader, directory) {
			reading++;

			reader.readEntries(function (entries) {
				reading--;
				if (entries.length > 0) {
					for (const entry of entries) {
						readEntry(entry, `${directory}/`);
					}

					readReaderContent(reader, `${directory}/`);
				}

				if (reading === 0) {
					resolve(contents);
				}
			});
		}
	});
}
