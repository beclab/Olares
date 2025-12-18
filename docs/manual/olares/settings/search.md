---
outline: [2,3]
description: Configure file indexing in Olares to control search scope, manage exclusions, and enable full-text search for specific directories.
---
# Configure file search

File search settings allow you to customize how Olares indexes and searches your data. By defining specific rules, you can control which files are excluded from global search and which directories support deep content searching.

This guide helps you understand and configure indexing rules to optimize your search experience.

:::tip Search scope 
These settings apply only to your own files in Olares.

Files or folders shared with you by other team members are searchable globally by default and do not need to be manually added to these rules. 
:::

## Understand search settings

Olares file search consists of three key components. Understanding how they interact helps you manage your system's indexing behavior.

### Search index

The **Search Index** section shows the current indexing status.

**Status meanings**
- `In Progress`: Olares is updating the index, so search results may be temporarily incomplete.
- `Completed`: Indexing is finished, and search results should be up to date.

**What can trigger `In Progress`**
- **File changes**: Files or folders are added, deleted, or updated in Olares.
- **Rule changes**: Excluded Files patterns change, or the Full-Text Search Directories list changes.

**How changes are applied**
- **Auto-update**: After file changes or rule changes, Olares updates the index automatically within about 30 seconds.
- **Rebuild**: If you want changes to take effect immediately (instead of waiting), click **Rebuild**.

In most cases, you can simply wait for the automatic update. **Rebuild** is mainly useful when you need results to reflect changes right away.

### Excluded files

This setting acts as a global ignore list for the search engine.

- Files matching these patterns are completely excluded from the index.
- Excluded files do not appear in search results, either by filename or by content.
- Rules are defined using regular expressions (regex).

### Full-text search directories

By default, Olares indexes files by filename only. This setting allows you to enable content search for specific folders.

- Files in listed directories are searchable by both filename and their text content.
- Files in unlisted directories remain searchable only by their filename.

## View search index status

To view the current search indexing status:

1. Open Settings from the Launchpad.
2. Go to **Search** > **File Search**.
3. Locate the Search Index section at the top of the panel.

If you need changes reflected immediately, click **Rebuild**.

## Configure search rules

On the File Search page, you can optimize results by excluding irrelevant files from indexing and by enabling full-text search for specific folders.

### Exclude files from search

Use the Excluded Files section if certain files (such as logs or temporary files) clutter your search results.

You exclude them by adding exclusion rules. Each rule uses a **regex pattern** to specify the file paths you want Olares to ignore.

**To add a rule:**
1. Under Excluded Files, click **Add pattern**.
2. Enter a valid regular expression (regex) to match the files you wish to ignore.
3. Click **Confirm**.

**To remove a rule:**
1. Under Excluded Files, locate the target rule.
2. Click the <i class="material-symbols-outlined">delete</i> icon next to the pattern.
3. In the prompt window, click **Confirm**.

:::warning Regex patterns
If you are not familiar with regular expressions, start with simple patterns and test them carefully. Incorrect patterns may cause important files to disappear from search results
:::

### Search inside file content

Use the Full-Text Search Directories section when you want to search inside documents, not just by filename.

By adding a **folder** to the full-text directory list, full-text search is enabled. Once added, supported files in that folder become searchable by both filename and content.

**To add a rule:**
1. Under Full-Text Search Directories, click **Add directory**.
2. Select the target folder from the file picker.
3. Click **Confirm** to add it to the list.

**To remove a rule:**
1. Under Full-Text Search Directories, locate the target directory.
2. Click the <i class="material-symbols-outlined">delete</i> icon next to the directory.
3. In the prompt window, click **Confirm**.

:::tip Supported formats 
Full-text search is supported for the following file types:<br> `.pdf`, `.doc`, `.docx`, `.csv`, `.rtf`, `.txt`, `.md`, `.json`, `.xml`.
:::

#### [Search within Olares](/manual/olares/desktop.md#search-within-olares)

Learn how to search within Olares.