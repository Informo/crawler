// Copyright 2018 Informo core team <core@informo.network>
//
// Licensed under the GNU Affero General Public License, Version 3.0
// (the "License"); you may not use this file except in compliance with the
// License.
// You may obtain a copy of the License at
//
//     https://www.gnu.org/licenses/agpl-3.0.html
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"regexp"
	"time"

	"gopkg.in/yaml.v2"
)

// FeedType represents the type of the feed: either RSS or Atom.
type FeedType int

// The various kind of feed types
const (
	FeedTypeRSS = iota
	FeedTypeAtom
)

// Config represents the overall architecture of the configuration file.
type Config struct {
	Crawler     CrawlerConfig  `yaml:"crawler"`
	Websites    []*Website     `yaml:"websites"`
	Database    DatabaseConfig `yaml:"database"`
	FeedsConfig FeedsConfig    `yaml:"feeds"`
}

// CrawlerConfig represents the specific configuration for the crawler, which
// will be applied across all instances.
type CrawlerConfig struct {
	UserAgent  string        `yaml:"user_agent"`
	RobotAgent string        `yaml:"robot_agent"`
	CrawlDelay time.Duration `yaml:"crawl_delay"`
}

// Website represents the configuration needed to describe a website a crawler
// will explore.
type Website struct {
	Identifier  string       `yaml:"identifier"`
	StartPoint  string       `yaml:"start_point"`
	Selectors   CSSSelectors `yaml:"selectors"`
	DateFormat  string       `yaml:"date_format"`
	MaxVisits   int          `yaml:"max_visits,omitempty"`
	IgnoreQuery bool         `yaml:"ignore_query,omitempty"`
	Filters     CrawlFilters `yaml:"filters,omitempty"`
}

// CrawlFilters represents the filters to apply when crawling a website.
type CrawlFilters struct {
	Restrict *regexp.Regexp `yaml:"restrict,omitempty"`
	Exclude  *regexp.Regexp `yaml:"exclude,omitempty"`
}

// UnmarshalYAML parses the regexps specified as filters and prepare them to be
// used when filtering the crawlers queues.
// Returns an error if there was an issue parsing the YAML source or parsing one
// of the regexps.
func (c *CrawlFilters) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		Restrict string `yaml:"restrict,omitempty"`
		Exclude  string `yaml:"exclude,omitempty"`
	}

	if err := unmarshal(&cfg); err != nil {
		return err
	}

	// All filters are optional, so we need to make sure one has been filled
	// before trying to parse it.
	if len(cfg.Restrict) > 0 {
		restrict, err := regexp.Compile(cfg.Restrict)
		if err != nil {
			return err
		}
		c.Restrict = restrict
	}
	if len(cfg.Exclude) > 0 {
		exclude, err := regexp.Compile(cfg.Exclude)
		if err != nil {
			return err
		}
		c.Exclude = exclude
	}

	return nil
}

// CSSSelectors represents the CSS selectors used to locate the different
// elements of a news item in a page.
type CSSSelectors struct {
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Content     string `yaml:"content"`
	Author      string `yaml:"author,omitempty"`
	Date        string `yaml:"date"`
	Thumbnail   string `yaml:"thumbnail,omitempty"`
}

// DatabaseConfig represents the needed configuration to talk to the database.
// There's two supported drivers: "postgres" and "sqlite3".
type DatabaseConfig struct {
	DriverName     string `yaml:"driver"`
	ConnectionData string `yaml:"connection_data"`
}

// FeedsConfig represents the configuration of the feeds exposed by the RSS
// generator.
type FeedsConfig struct {
	Type FeedType
}

// UnmarshalYAML detects the type of feed and sets the right value into the
// FeedsConfig instance.
// Returns an error if decoding the YAML configuration failed, or if the provided
// type is invalid (ie neither "rss" nor "atom").
func (fc *FeedsConfig) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var cfg struct {
		Type string `yaml:"type"`
	}

	if err := unmarshal(&cfg); err != nil {
		return err
	}

	switch cfg.Type {
	case "rss":
		fc.Type = FeedTypeRSS
		break
	case "atom":
		fc.Type = FeedTypeAtom
		break
	default:
		return fmt.Errorf("Invalid feed type: %s", cfg.Type)
	}

	return nil
}

// Load reads the configuration file located at a given path and loads it into
// a Config instance.
// Returns an error if there was an issue reading the file, parsing it, or if
// one of the website's URL is invalid.
func Load(filePath string) (cfg *Config, err error) {
	// Reads the configuration file.
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}

	// Parse the configuration file into the Config instance.
	cfg = new(Config)
	if err = yaml.Unmarshal(content, cfg); err != nil {
		return
	}

	// Check the URL provided for each website.
	for _, w := range cfg.Websites {
		if _, err = url.Parse(w.StartPoint); err != nil {
			err = fmt.Errorf(
				"Start point for %s isn't a valid URL: %s", w.Identifier, err.Error(),
			)
			return
		}
	}

	// Check if the database driver is supported.
	if cfg.Database.DriverName != "postgres" && cfg.Database.DriverName != "sqlite3" {
		err = fmt.Errorf("Unsupported database driver %s", cfg.Database.DriverName)
		return
	}

	return
}
