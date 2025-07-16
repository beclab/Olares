# `olares release`

## Synopsis
Build a release version based on a local Olares repository. This command should be run in the root directory of the Olares repository.

```bash
olares-cli olares release [option]
```

## Options

| Name                       | Shorthand | Usage                                                                                                                                                                                                                                                   |
|----------------------------|-----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `--base-dir`               | `-b`      | Sets the base directory for Olares.<br> Defaults to `$HOME/.olares`.                                                                                                                                                                                |
| `--download-cdn-url`       |           | Sets the CDN URL used for downloading checksums of dependencies and images. <br> Defaults to `https://dc3p1870nn3cj.cloudfront.net`.                                                                                                                    |
| `--extract`                | `-e`      | Extracts the release to the `--base-dir` after the build. Set to `false` if only the release file itself is needed. <br> Defaults to `true`.                                                                                                            |
| `--help`                   | `-h`      | Displays help information.                                                                                                                                                                                                                              |
| `---ignore-missing-images` |           | Ignores missing images when downloading checksums from the CDN. <br> Disable this only if no new images are added, as the build may fail if the image is not uploaded to the CDN yet. <br> Defaults to `true`.                                          |
| `--version`                | `-v`      | Specifies the Olares version. <br>Version values follow the format `x.y.z` (e.g., `1.10.0`) or include a build date (e.g., `1.10.0-20241109`).<br> Refer to the [GitHub Releases page](https://github.com/beclab/Olares/releases) for available versions. |

