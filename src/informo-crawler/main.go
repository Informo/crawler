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

package main

import (
	"flag"
	"fmt"
	"sync"

	"common/config"
	"common/database"
	"informo-crawler/crawler"

	"github.com/sirupsen/logrus"
)

var (
	configFile = flag.String("config", "config.yaml", "Configuration file")
	debug      = flag.Bool("debug", false, "Print debugging messages")
)

func main() {
	// Parse the command line arguments.
	flag.Parse()

	// Configure the logger.
	logConfig()

	// Load the configuration from the provided configuration file.
	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't load config: %s", err.Error()))
	}

	// Open the database and prepare the required statements.
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't open database: %s", err.Error()))
	}

	// Using a sync.WaitGroup to keep track of the goroutines and only exit when
	// all goroutines have returned.
	var wg sync.WaitGroup
	// Storing the crawlers in a slice outside of the loop and the goroutines in
	// case we need to use it later.
	crawlers := make([]*crawler.Crawler, len(cfg.Websites))
	// Spawn a crawler for each website.
	for i, w := range cfg.Websites {
		// Instantiate the crawler.
		crawlers[i] = crawler.NewCrawler(cfg.Crawler, db, w)

		// Increment the sync.WaitGroup's counter.
		wg.Add(1)

		// Run the crawler in a separate goroutine.
		go func(c *crawler.Crawler) {
			// Run the crawler and, when it stops, retrieve the reason that made
			// it stop if there's one.
			stopReason := c.Run()

			// If a reason was provided, log it, if not, don't.
			if len(stopReason) > 0 {
				c.Log.Info(fmt.Errorf("Crawler stopped: %s", stopReason))
			} else {
				c.Log.Warn(fmt.Errorf("Crawler stopped"))
			}

			// Tell the sync.WaitGroup that the goroutine has ended.
			wg.Done()
		}(crawlers[i])
	}

	// Wait for all goroutines to end before exiting.
	wg.Wait()
}
