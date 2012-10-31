//
// Blackblog
// Copyright 2012 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package main

import (
	"flag"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	flagSource    = flag.String("root", "", "The root directory of all Markdown posts")
	flagDest      = flag.String("dest", "", "The output directory for running in comiple mode")
	flagTemplates = flag.String("templates", "templates/", "The directory containing the Blackblog templates")

	commandDocs = map[string]string{
		"newblog": "Create a new blog with some sample data in the specified directory.",
		"serve":   "Run a standalone web server for the given blog.",
		"render":  "Render the blog out to static HTML files.",
	}
	commandOrder = []string{
		"newblog",
		"serve",
		"render",
	}
)

func main() {
	flag.Usage = usage
	flag.Parse()
	args := flag.Args()

	if len(args) != 1 {
		usage()
		os.Exit(1)
	}

	if *flagSource == "" {
		fmt.Fprintf(os.Stderr, "No -root blog directory specified\n")
		os.Exit(1)
	}

	if !RunAsServer() && *flagDest == "" {
		fmt.Fprintf(os.Stderr, "No -port or -dest flag specified\n")
		os.Exit(2)
	}

	if RunAsServer() {
		if err := StartBlogServer(*flagSource); err != nil {
			fmt.Fprint(os.Stderr, "Could not start blog server:", err)
			os.Exit(3)
		}
	}

	posts, err := GetPostsInDirectory(*flagSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "GetPostsInDirectory: %v\n", err)
		os.Exit(3)
	}

	renderTree, err := createRenderTree(posts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "createRenderTree: %v\n", err)
		os.Exit(3)
	}

	if err := writeRenderTree(*flagDest, renderTree); err != nil {
		fmt.Fprintf(os.Stderr, "writeRenderTree: %v\n", err)
		os.Exit(3)
	}

	index, err := CreateIndex(posts)
	var f *os.File
	if err == nil {
		f, err = os.Create(path.Join(*flagDest, "index.html"))
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "writing index: %v\n", err)
		os.Exit(3)
	}
	defer f.Close()
	f.Write(index)
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  Commands:\n")
	for _, cmdName := range commandOrder {
		fmt.Fprintf(os.Stderr, "    %s\t%s\n", cmdName, commandDocs[cmdName])
	}
	fmt.Fprintf(os.Stderr, "\n  Flags:\n")
	flag.PrintDefaults()
}

// GetPostsInDirectory recursively examines the directory at the path and finds
// any Markdown (.md) files and returns the corresponding Post objects.
func GetPostsInDirectory(dirPath string) (posts PostList, err error) {
	err = filepath.Walk(dirPath, func(file string, info os.FileInfo, err error) error {
		if !info.IsDir() && strings.HasSuffix(file, ".md") {
			if post, err := NewPostFromPath(file); post != nil && err == nil {
				posts = append(posts, post)
			} else {
				return err
			}
		}
		return nil
	})
	return
}

func getRootPath(name string) string {
	return strings.Repeat("../", strings.Count(name, "/"))
}
