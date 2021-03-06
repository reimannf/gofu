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

package prompt

import (
	"os"
	"path/filepath"
	"strings"
)

//Directory contains all data about a directory that the prompt needs.
type Directory struct {
	Path                  string
	DisplayPath           string
	InBuildTree           bool
	InRepoTree            bool
	RepoRootPath          string
	NearestAccessiblePath string
}

//CurrentDirectory prepares a Directory struct for the current working
//directory.
func CurrentDirectory() Directory {
	cwd, err := os.Getwd()
	if err != nil {
		cwd = filepath.Clean(os.Getenv("PWD"))
	}
	return NewDirectory(cwd)
}

//NewDirectory prepares a Directory struct for the given path.
func NewDirectory(path string) (dir Directory) {
	dir.Path = path
	dir.DisplayPath = path

	//check if path actually exists
	dir.NearestAccessiblePath = findNearestAccessiblePath(dir.Path)
	if dir.NearestAccessiblePath == dir.Path {
		//marks that everything is okay existence-wise
		dir.NearestAccessiblePath = ""

		//display tag if below /x/build
		if buildPath := os.Getenv("BUILD_ROOT"); buildPath != "" {
			rel, _ := filepath.Rel(buildPath, dir.DisplayPath)
			if !strings.HasPrefix(rel, "..") && rel != "." {
				dir.InBuildTree = true
				dir.DisplayPath = filepath.Join("/", rel)
			}
		}

		//display tag if below /x/src
		if gopath := os.Getenv("GOPATH"); gopath != "" {
			repoPath := filepath.Join(gopath, "src")
			rel, _ := filepath.Rel(repoPath, dir.DisplayPath)
			if !strings.HasPrefix(rel, "..") && rel != "." {
				dir.InRepoTree = true
				dir.DisplayPath = rel
			}
		}

		//strip $HOME prefix if applicable and desirable
		dir.stripHomeDirFromDisplay()

		//check if we are inside a Git repository
		dir.RepoRootPath = findRepoRootPath(dir.Path)
	}

	return
}

//This part can benefit from a "return" in the middle, so it's in a separate function.
func (dir *Directory) stripHomeDirFromDisplay() {
	if !strings.HasPrefix(dir.DisplayPath, "/") {
		return
	}

	homePath := os.Getenv("HOME")
	if homePath == "" {
		return
	}

	rel, _ := filepath.Rel(homePath, dir.DisplayPath)
	if rel == "." {
		//do not display an empty DisplayPath if tags are displayed
		if dir.InBuildTree || dir.InRepoTree {
			return
		}
		rel = ""
	}
	if !strings.HasPrefix(rel, "..") {
		dir.DisplayPath = rel
	}
}

func findNearestAccessiblePath(path string) string {
	_, err := os.Stat(path)
	if err == nil {
		return path
	}
	return findNearestAccessiblePath(filepath.Dir(path))
}

func getDirectoryField(dir Directory) string {
	if dir.DisplayPath == "" {
		return ""
	}

	txt := withColor("1;36", dir.DisplayPath)
	if dir.NearestAccessiblePath == "" {
		//cwd accessible -> highlight path elements inside the repo (if any)
		if dir.RepoRootPath != "" && dir.RepoRootPath != dir.Path {
			rel, _ := filepath.Rel(dir.RepoRootPath, dir.Path)
			if strings.HasSuffix(dir.DisplayPath, rel) {
				base := strings.TrimSuffix(dir.DisplayPath, rel)
				txt = withColor("0;36", base) + withColor("1;36", rel)
			}
		}
	} else {
		//cwd inaccessible -> highlight inaccessible path elements
		rel, _ := filepath.Rel(dir.NearestAccessiblePath, dir.Path)
		txt = withColor("1;36", dir.NearestAccessiblePath+"/") + withColor("1;31", rel)
	}

	//apply tags
	if dir.InRepoTree {
		txt = withType("repo", txt)
	}
	if dir.InBuildTree {
		txt = withType("build", txt)
	}
	return txt
}

func getDeletedMessageField(dir Directory) string {
	if dir.NearestAccessiblePath == "" {
		return ""
	}
	return withColor("1;41", "cannot stat cwd")
}
