---
outline: [2,4]
description: Explain key syntax rules and available filters in Wise.
---
# Filter syntax reference

Filtered views are powerful tools for organizing your entries. You can customize your queries using various parameters and operators. Here's a breakdown of the key syntax rules and available filters.

## Query structure
To create a query with a single condition, simply specify the parameter and its value.

**Example**:
```
feed_id:12345
```

You can also combine multiple parameters using logical operators.

### `AND`
Use `AND` to link parameters that must *all apply* to the results.  

**Example**:

Retrieve entries that are unread *and* marked as "Read later":
```
location:readlater AND seen:false
```

### `OR`
Use `OR` to link parameters where *at least one* condition must apply.

**Example**:

Retrieve entries that are from the library *or* the last opened date is earlier than January 1, 2023:
```
islibrary:true OR last_opened__lt:2023-01-01
```

### Group with parentheses
You can use parentheses (`()`) to group conditions and control the order of evaluation in your queries.

**Example**:

Retrieve entries that are unread *and* either in the inbox *or* marked as "Read Later":
```
seen:false AND (location:inbox OR location:readlater)
```

### Special characters
If a filter condition contains special characters (such as spaces, colons, etc.), enclose it in double quotes `""` (excluding the double quotes themselves). For example:
```
tag:"Project/AI"
```
```
author:"Arthur C. Clarke"
```
## Parameters

### Basic parameters
Use the format `parameter:value` to filter entries based on specific conditions.

#### `feed_id`
Feed ID the entry belongs to.
You can find the Feed ID in the <i class="material-symbols-outlined">settings</i> > **RSS feeds** page.

#### `author`
Author of the entry.

#### `file_type`
The file type of the entry.
Valid file types are:
- `article`
- `video`
- `audio`
- `ebook`
- `pdf`

### Time parameters
Filter entries based on time-related fields using the format `parameter__operator:value`. Dates must be specified in absolute format: `YYYY-MM-DD`.

#### `published_at`
Publication time of the entry.

#### `created_at`
Creation time of the entry.

#### `updated_at`
Last updated time of the entry.

#### `last_opened`
Last time the entry was opened.

#### Operators
Available operations are:
- `__gt` (Greater than)<br/>
  Retrieve entries where the time is **after** the specified date.

- `__gte` (Greater than or equal to)<br/>
  Retrieve entries where the time is **on or after** the specified date.

- `__lt` (Less than)<br/>
  Retrieve entries where the time is **before** the specified date.

- `__lte` (Less than or equal to)<br/>
  Retrieve entries where the time is **on or before** the specified date.

### Boolean parameters
Filter entries based on true/false values using the format `field_name:true/false`.

#### `islibrary`
Whether the entry is from a "library" source.

#### `isfeed`
Whether the entry is from a "feed" source.

#### `seen`
Whether the entry has been read.

### Location parameters
Filter entries based on their location.

#### `location`
Indicates the location of the entry.

Available values are:
- `all`: All entries.
- `readlater`: Entries marked as "Read Later".
- `inbox`: Entries in the inbox.

### Related content parameters
Filter entries based on whether they have specific related content.

#### `has`
Use the `has` parameter to filter entries by the presence of specific content.

Available values are:
- `note`: Whether the entry has notes.
- `tag`: Whether the entry has tags.

#### `tag`
Filter entries by the name of a specific tag.

#### `tag_id`
Filter entries by the unique ID of a specific tag.
You can find the Feed ID in the <i class="material-symbols-outlined">settings</i> > **Tags** page.