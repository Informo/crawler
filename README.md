# Informo extractor

The Informo extractor is a tool that allows extracting some content from a news website, and generating a feed with the extracted content.

It contains two separate components: the Informo crawler and the Informo feed generator.

## Informo crawler

The Informo crawler is a programm that visits given websites, parse each page and, for each page it sees as a news item, extract the item's content and metadata (time of publication, author, title, etc.) and saves them in the database. More than one website can be specified in the configuration, and, when started, it will start visiting all of them in parallel.

The crawler stops once all website have been entirely visited, so it isn't designed to be used as a daemon, but rather as a recurrent task.

## Informo feed generator

The Informo feed generator is a lightweight Web server using the data extracted from the crawler to generate a feed, which can be either a RSS or an Atom feed (depending on the configuration). The server will run on the given interface and port, and will handle all `GET` requests to `/website`, where `website` is the identifier of the source website, as it appears in the configuration file. Generated feeds are compatible with the [Informo feeder](https://github.com/Informo/informo-feeder).

## Build

You can either install the Informo extractor by using a release on one of the [repository's releases](https://github.com/Informo/informo-extractor/releases), or by building it by yourself.

The projet is built using [gb](https://getgb.io/), which you can install by running:

```
go get github.com/constabulary/gb/...
```

Then all you need to do is:

```bash
git clone https://github.com/Informo/informo-extractor
cd informo-extractor
gb build
```

You can the run the Informo crawler by calling:

```
./bin/informo-crawler
```

and the Informo feed generator by calling:

```
./bin/informo-feed-generator
```

## Configuration

Most of the configuration keys are already widely documented with examples in the `config.sample.yaml` file, so this section won't say much about them. The same configuration file is used for both the crawler and the feed generator.

### Date format

When configuring a website to be visited by the crawler, you are required to input the format the articles' dates follow when displayed on the website. This allows the crawler to decode the date in a way it can understand.

The format consists of a string where the different elements used to describe the date and the time are replaced with patterns (see below for examples). The available patterns are the following:

* **`{DAY_LONG}`** is the long form of the day's name (e.g. "Monday")
* **`{DAY_SHORT}`** is the short form of the day's name (e.g. "Mon")
* **`{DAY_NUM}`** is the number of the day in the month (e.g. "2")
* **`{MONTH_LONG}`** is the long form of the month's name (e.g. "January")
* **`{MONTH_SHORT}`** is the short form of the month's name (e.g. "Jan")
* **`{MONTH_NUM}`** is the number of the month in the year (e.g. "1")
* **`{YEAR_LONG}`** is the long form of the year (e.g. "2006")
* **`{YEAR_SHORT}`** is the short form of the year (e.g. "06")
* **`{HOURS}`** is the time's hours (e.g. "15")
* **`{MINUTES}`** is the time's minutes (e.g. "04")
* **`{SECONDS}`** is the time's seconds (e.g. "05")
* **`{ZONE_OFFSET}`** is the time zone's offset (e.g. "-0700")
* **`{ZONE_ABBREV}`** is the time zone's abbreviation (e.g. "MST")

It is, of course, not mandatory to include all patterns in the format. Please note that `{DAY_NUM}` and `{MONTH_NUM}` can start with a "0" or not, it doesn't matter.

#### Examples

All examples below are using January 2nd, 2006 as the example date.

* The time format matching "2 January 2006" would be `{DAY_NUM} {MONTH_LONG} {YEAR_LONG}`
* The time format matching "01/02/2006" would be `{MONTH_NUM}/{DAY_NUM}/{YEAR_LONG}`
* The time format matching "Monday 2 January at 15:04" would be `{DAY_LONG} {DAY_NUM} {MONTH_LONG} at {HOURS}:{MINUTES}`

## Repo structure

All the Informo extractor's code is located under the `src` directory. It is split in three parts:

* `informo-crawler` contains the code for the crawler
* `informo-feed-generator` contains the code for the feed generator
* `common` contains the code used by both components (e.g. configuration loading, database access)
