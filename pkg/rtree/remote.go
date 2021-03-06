/*******************************************************************************
*
* Copyright 2017 Stefan Majewsky <majewsky@gmx.net>
*
* This program is free software: you can redistribute it and/or modify it under
* the terms of the GNU General Public License as published by the Free Software
* Foundation, either version 3 of the License, or (at your option) any later
* version.
*
* This program is distributed in the hope that it will be useful, but WITHOUT ANY
* WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
* A PARTICULAR PURPOSE. See the GNU General Public License for more details.
*
* You should have received a copy of the GNU General Public License along with
* this program. If not, see <http://www.gnu.org/licenses/>.
*
*******************************************************************************/

package rtree

import (
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

//ExpandRemoteURL derive the canonical URL for a given remote by substituting
//aliases defined in the system-wide and user-global Git config. For example,
//with
//
//    $ cat /etc/gitconfig
//    [url "git://github.com/"]
//    insteadOf = gh:
//
//and the input "gh:foo/bar", this function returns "git://github.com/foo/bar".
func ExpandRemoteURL(remoteURL string) string {
	var best *RemoteAlias
	for _, current := range RemoteAliases {
		if strings.HasPrefix(remoteURL, current.Alias) {
			if best == nil || len(best.Alias) < len(current.Alias) {
				best = current
			}
		}
	}
	if best == nil {
		return remoteURL
	}
	return best.Replacement + strings.TrimPrefix(remoteURL, best.Alias)
}

//This regex recognizes the scp-like syntax for git remotes
//(i.e. "[user@]example.org:path/to/repo") as specified by the "GIT URLS"
//section of man:git-clone(1).
var scpSyntaxRx = regexp.MustCompile(`^(?:[^/@:]+@)?([^/:]+\.[^/:]+):(.+)$`)

//CheckoutPathForRemoteURL derives the checkout path for a remote URL that has
//already been expanded with ExpandRemoteURL() if necessary.
//
//  "https://example.org/foo/bar" -> "example.org/foo/bar"
//  "git@example.org:foo/bar"     -> "example.org/foo/bar"
//
func CheckoutPathForRemoteURL(remoteURL string) (string, error) {
	match := scpSyntaxRx.FindStringSubmatch(remoteURL)
	if match != nil {
		//match[1] is the hostname, match[2] is the path to the repo
		return filepath.Join(match[1], match[2]), nil
	}

	u, err := url.Parse(remoteURL)
	if err != nil {
		return "", err
	}
	return filepath.Join(u.Hostname(), u.Path), nil
}
