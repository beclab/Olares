---
description: Learn how to use advanced filters in Wise to organize, prioritize, and rediscover your saved content. These examples cover practical use cases like focusing on unread entries and revisiting old content.
---

# Filtered view examples
This guide provides examples of filter queries for Wise. Each example addresses a specific use case, making it easy to organize, sort, and rediscover content in your library.

## Focus on unread entries from a specific feed
Want to catch up on just one feed? Use this query to display all unread entries from a particular source.
```
feed_id:12345 AND seen:false
```

## See everything by your favorite author
Easily pull up all entries written by a specific author you love.
```
author:"John Doe"
```
This shows all documents where the author is "John Doe" so you can stay up-to-date with a writer's work.

## Stay up-to-date with the latest content
Only want to see new articles? Filter for content published after a specific date.
```
published_at__gt:2025-04-01
```
This query displays entries published after April 1, 2025. It's a great way to focus on fresh updates and avoid getting lost in older material.

## Review saved content with notes
Use this query to find tagged entries that include your own notes:
```
tag:AI AND has:note
```
This is useful for revisiting saved content where you've added personal insights or comments.

## Prioritize tagged "Read Later" entries
Need to focus on specific topics? This filter shows all your "Read Later" entries with a particular tag.
```
tag:Work AND location:readlater
```
This pulls up all entries tagged with "Work" that are saved to your "Read Later" list. It's perfect for tackling important projects or topics you've set aside.

## Clean up your inbox with unread videos
Want to catch up on videos from your RSS feeds? This filter focuses on unwatched content.

```
file_type:video AND seen:false AND isfeed:true
```
This shows all unread video entries from your subscribed feeds. It's ideal for binge-watching or sorting through your video backlog.

## Revisit entries you haven't opened in months
Got old productivity-related entries that you haven't touched in a while? This filter helps you rediscover them.

```
last_opened__lt:2023-12-31 AND tag:Productivity
```
This shows all entries that you last opened before December 31, 2023. You can revisit useful content you've saved for productivity but might have forgotten about—or decide if it's time to delete it.