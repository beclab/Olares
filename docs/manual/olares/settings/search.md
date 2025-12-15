---
outline: [2,3]
description: Configure file indexing in Olares to control search scope, manage exclusions, and enable full-text search for specific directories.
---
# Configure file search

File Search settings allow you to customize how Olares indexes and searches your data. By defining specific rules, you can control which files are excluded from global search and which directories support deep content searching.

This guide helps you understand and configure indexing rules to optimize your search experience.

:::tip Search scope 
These settings apply only to your own files in Olares.

Files or folders shared with you by other team members are searchable globally by default and do not need to be manually added to these rules. 
:::

## Understand search settings

Olares file search consists of three key components. Understanding how they interact helps you manage your system's indexing behavior.

### Search index
The **Search Index** section lets you control how and when indexing changes are applied.

- **Auto-update**: When you change search rules, Olares automatically updates the index within approximately 30 seconds.
- **Manual rebuild**: You can click **Rebuild** to apply the latest search rules immediately, instead of waiting for the automatic update.

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

To view your current search index status:

1. Open Settings from the Launchpad.
2. Go to **Search** > **File Search**.
3. Locate the Search Index section at the top of the panel.
   - If files in your Olares are being actively created, deleted, or updated, the status shows as `In Progress`.
   - If files in your Olares remain unchanged and indexing is complete, the status shows as `Completed`.

:::tip When to manually rebuild 
While Olares applies rules automatically, you may want to click **Rebuild** to apply changes immediately when:
- You have just added or removed a Full-Text Search Directory.
- You have updated Excluded Files patterns.
:::

## Configure search rules

On the File Search page, you can manage your exclusion patterns and full-text scopes.

### Excluded files

Use the Excluded Files section to prevent specific files (like system logs or temporary files) from cluttering your search results.

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

### Full-text search directories

Use the Full-Text Search Directories section to specify where Olares should look inside your files.

**To add a rule:**
1. Under Full-Text Search Directories, click **Add directory**.
2. Select the target folder from the file picker.
3. Click **Confirm** to add it to the list.

**To remove a rule:**
1. Under Full-Text Search Directories, locate the target directory.
2. Click the <i class="material-symbols-outlined">delete</i> icon next to the directory.
3. In the prompt window, click **Confirm**.

:::tip Supported formats 
Full-text search is supported for the following file types:<br> `.pdf`, `.doc`, `docx`, `.csv`, `.rtf`, `.txt`, `.md`, `.`json`, `.xml` 
:::

#### [Search within Olares](/manual/olares/desktop.md#search-within-olares)

Learn how to search within Olares.