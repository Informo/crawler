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
	"strings"
)

// patterns contains the list of known patterns used in date layouts. The name
// of each pattern is written without the curly brackets, which are added in the
// replaceLayoutPatterns function.
var patterns = map[string]string{
	"DAY_LONG":    "Monday",
	"DAY_SHORT":   "Mon",
	"DAY_NUM":     "2",
	"MONTH_LONG":  "January",
	"MONTH_SHORT": "Jan",
	"MONTH_NUM":   "1",
	"YEAR_LONG":   "2006",
	"YEAR_SHORT":  "06",
	"HOURS":       "15",
	"MINUTES":     "04",
	"SECONDS":     "05",
	"ZONE_OFFSET": "-0700",
	"ZONE_ABBREV": "MST",
}

// replaceLayoutPatterns replaces all known {PATTERN}s in the date layout (aka
// date format) with the correct values so it can be read by time.Parse().
// Using patterns (which we replace at startup) has two major benefits for the
// end user:
//   * they don't have to use January 2nd, 2006 as reference, which reason isn't
//     always obvious, and therefore can induce incomprehensions, which usually
//     results in a bad user experience.
//   * we currently use the "monday" library to parse non-English dates, which
//     requires layouts to be written in English. This can be hard to understand
//     for the user (because the obvious thing to do would be writing the layout
//     in the same language as the date), and it might induce mixes between the
//     foreign language and English in the layout. Although the patterns are also
//     written in english, their syntax makes more it more obvious to notice that
//     they are actually placeholders, whereas it's not that much obvious for
//     words such as "Monday" or "January".
// Doesn't return any error, and replaces every occurrence of every {PATTERN} in
// the layout string.
func replaceLayoutPatterns(layout *string) {
	// Iterate over each pattern and its replacement string.
	for pattern, replacement := range patterns {
		// We call strings.Replace() with n = -1 so it replaces every occurrence
		// of the {PATTERN}, and not a limited number of these.
		*layout = strings.Replace(*layout, fmt.Sprintf("{%s}", pattern), replacement, -1)
	}
}
