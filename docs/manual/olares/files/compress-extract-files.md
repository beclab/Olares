---
outline: [2, 3]
description: Compress and extract files in the Olares Files app. Supports ZIP, 7z, TAR, and more formats, with advanced options for compression level, password protection, and split archives.
---

# Compress and extract files

In the Files app, you can compress files to save space or share them, and extract archives in various common formats.

## Compress files

1. Open the Files app.
2. Right-click the file or folder you want to compress, and then hover over **Compress to...**.

    ![Compress files](/images/manual/olares/files-compress-to.png#bordered)

3. To compress it quickly, select **ZIP file**, **7z file**, or **TAR file**. The archive will be created directly in the current folder.
4. To compress to other formats or customize compression settings, select **More options**:

    a. In the **Create archive** window, configure the following settings as needed:

    ![Create archive settings](/images/manual/olares/files-compress.png#bordered){width=70%}

    | Parameter | Description |
    |:----|:----|
    | **Archive name** | Enter a name for the archive. |
    | **Save to** | Choose the destination path for the archive. |
    | **Format** | Select the compression format: **ZIP**, **7z**, **TAR**, **tar.gz**, **tgz**, **tar.bz2**, **tar.xz**,<br>**gzip**, **bzip2**, or **xz**.<br><br>Note that **TAR**, **tar.gz**, **tgz**, **tar.bz2**, **tar.xz**, **gzip**, **bzip2**, and **xz** do not<br> support password settings, filename encryption, or split into volumes. |
    | **Compression level** | Drag the slider or enter a value between 1 and 9:<ul><li>A lower value means faster compression but a larger archive size.</li><li>A higher value means slower compression but a better compression ratio.</li></ul> |
    | **Password (optional)** | Set an encryption password for the archive. <br>This password is required when extracting the archive. |
    | **Confirm password** | Re-enter the password to confirm. |
    | **Split into volumes (optional)** | Set the maximum size for each volume and select the unit (KB, MB, or GB).<br>If the resulting archive exceeds this limit, it will be automatically split into<br> multiple volumes, each within the set size.<br><br>Use this feature to split large archives so they fit within file size limits <br>for uploads such as email attachments or cloud storage.<br><br>Note: When extracting, all volume files must be in the same folder.|
    | **On conflict** | Choose how to handle the situation when a file with the same name<br> already exists in the destination folder:<ul><li>**Rename (add suffix)**: Keep the existing file and automatically add <br>a sequence number suffix to the new archive name.<br>For example, `file(1).zip`.</li><li>**Overwrite**: Replace the existing file with the newly created archive.</li><li>**Skip**: Skip the conflicting file and keep the existing file.</li></ul> |
    | **Preserve symbolic links** | Specify how to handle symbolic links (similar to shortcuts):<ul><li>If you select this option, only the link itself is saved in the archive, <br>taking up very little space.</li><li>If you do not select this option, the system will follow the link to the<br> actual source file and include it in the archive.</li></ul> |

    b. Click **Create**.

## Extract files

1. Open the Files app.
2. Right-click the target archive, and then hover over **Extract to...**.
3. To extract directly to the current location, select **Current folder**.

    ![Extract files](/images/manual/olares/files-extract-to.png#bordered)

4. To specify another path or customize extraction settings, click **Choose location**:

    a. In the **Extract archive** window, configure the following settings as needed:

    ![Extract archive settings](/images/manual/olares/files-extract.png#bordered){width=70%}

    | Parameter | Description |
    |:----|:----|
    | **Extract to** | Specify the destination path for the extracted files. |
    | **On conflict** | Choose how to handle the situation when an extracted file has the <br>same name as an existing file in the destination folder:<ul><li>**Rename (add suffix)**: Keep the existing file and automatically add a <br>sequence number suffix to the extracted file.</li><li>**Overwrite**: Replace the existing file with the extracted file.</li><li>**Skip**: Skip the conflicting file and keep the existing file.</li></ul> |
    | **Preserve symbolic links** | <ul><li>If you select this option, symbolic links in the archive will be <br>extracted as link files as they are.</li><li>If you do not select this option, they will be resolved into actual<br> files or folders during extraction.</li></ul> |
    | **Open folder when done** | If you want the system to automatically open the folder containing <br>the extracted content when extraction is completed, select this option. |

    b. Click **Extract**.
