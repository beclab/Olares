---
outline: [2,3]
description: Configure file indexing in Olares to control search scope, manage exclusions, and enable full-text search for specific directories.
---
# Configure file search

File search settings allow you to customize how Olares indexes and searches your data. By defining specific rules, you can control which files are excluded from global search and which directories support deep content searching.

This guide helps you understand and configure indexing rules to optimize your search experience.

See also [Search within Olares](/manual/olares/desktop.md#search-within-olares) for how to use global search.

:::tip Search scope 
These settings apply only to your own files in Olares.

Files or folders shared with you by other team members are searchable globally by default and do not need to be manually added to these rules. 
:::

## Understand search settings

Olares file search consists of three key components. Understanding how they interact helps you manage your system's indexing behavior.

![File search](/images/manual/olares/file-search.png#bordered)
### Search index

The **Search Index** section shows the current indexing status and lets you manually rebuild the index when needed.

After files change or you update search rules, Olares automatically refreshes the index within about 30 seconds. Most of the time, you can rely on this automatic update. Use **Rebuild** only when you need search results to reflect changes immediately.

The status indicator shows the current progress:

- **In Progress**: The index is being updated. Search results may be temporarily incomplete during this time.

- **Completed**: The index is up to date and search results should reflect the latest files and rules.

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