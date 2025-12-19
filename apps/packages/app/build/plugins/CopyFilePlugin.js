const fs = require('fs');
const path = require('path');

/**
 * webpack 脚本
 * @description 复制文件到指定目录
 */
module.exports = class CopyWebpackPlugin {
	/**
	 *
	 * @param { String } options[index].fromPath 文件来源，文件路径
	 * @param { String } options[index].fromName 文件来源，文件名称
	 * @param { String } options[index].toPath 复制文件路径
	 * @param { String } options[index].toName 复制文件名称
	 */
	constructor(options) {
		this.options = options;
	}

	// webpack 规定每个插件的实例，必须有一个 .apply() 方法，webpack 打包前会调用所有插件的方法，插件可以在该方法中进行钩子的注册。
	apply(compiler) {
		compiler.hooks.afterEmit.tapAsync(
			'CopyWebpackPlugin',
			(compilation, cb) => {
				try {
					if (!fs.copyFile) {
						console.error('Your nodejs version is too low, please upgrade!');
					} else {
						if (this.options.length) {
							this.options.forEach((option) => {
								const files = fs.readdirSync(option.fromPath, {
									withFileTypes: true
								});
								files.forEach((file) => {
									if (file.isFile() && file.name === option.fromName) {
										fs.copyFile(
											path.resolve(option.fromPath, file.name),
											path.resolve(option.toPath, option.toName),
											(error) => {
												if (error) {
													console.error(file.name + 'copy failed：' + error);
												}
											}
										);
									}
								});
							});
						}
					}
				} catch (error) {
					console.error(error);
				} finally {
					cb();
				}
			}
		);
	}
};
