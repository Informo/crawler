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
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the overall architecture of the configuration file.
type Config struct {
	Crawler  CrawlerConfig `yaml:"crawler"`
	Websites []*Website    `yaml:"websites"`
	Database string        `yaml:"database"`
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
	Identifier string       `yaml:"identifier"`
	StartPoint string       `yaml:"start_point"`
	Selectors  CSSSelectors `yaml:"selectors"`
	DateFormat string       `yaml:"date_format"`
	MaxVisits  int          `yaml:"max_visits,omitempty"`
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

	return
}
