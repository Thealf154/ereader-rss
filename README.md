# RSS Reader for ereaders

Being able to read current things in a ereader is pretty difficult since many ereaders don't handle well modern sites. EreaderRSS lets you see your rss feeds in your ereader.

## Docker deploy

Just run this command:

```bash
docker-compose up
```

## Features

- Cache system for rss feeds.
- No ads (can't be served in an ereader).
- All pages are server rendered.
- Legacy CSS so it can be supported in the ereaders experimental browsers.
- Download entire feed as an EPUB.
