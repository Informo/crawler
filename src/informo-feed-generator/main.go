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

	"common"
	"common/config"
	"common/database"
	"informo-feed-generator/generator"

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
	common.LogConfig(*debug)

	// Load the configuration from the provided configuration file.
	cfg, err := config.Load(*configFile)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't load config: %s", err.Error()))
	}

	// Check if there's a "feed" section in the configuration file.
	if cfg.FeedsConfig == nil {
		logrus.Panic(fmt.Errorf("No 'feeds' configuration found, please provide one"))
	}

	// Open the database and prepare the required statements.
	db, err := database.NewDatabase(cfg.Database)
	if err != nil {
		logrus.Panic(fmt.Errorf("Couldn't open database: %s", err.Error()))
	}

	// Instantiate the Generator.
	g := generator.NewGenerator(db, cfg.FeedsConfig)

	// Serve the feeds and listen on the configured interface and port.
	if err = g.SetupAndServe(); err != nil {
		logrus.Panic(fmt.Errorf("Something went wrong with the web server: %s", err.Error()))
	}
}
