// Copyright © 2020 Roman Dodin <dodin.roman@gmail.com>
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

package cmd

import (
	"fmt"
	"strings"

	path "github.com/hellt/yangpath/pkg/path"
	"github.com/openconfig/goyang/pkg/yang"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export xpath-styled paths from a given YANG module",

	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		if err = path.AddYANGDirs(viper.GetStringSlice("yang-dir")); err != nil {
			log.Fatal(err)
		}

		ms := yang.NewModules()
		if err := ms.Read(viper.GetString("module")); err != nil {
			log.Info(err)
		}
		if len(ms.Modules) == 0 {
			log.Fatal("no YANG modules found by the path specified.")
		}
		var mn string // module name
		// at this moment ms.Modules map contains either a single module
		// which was read by the path provided within -m flag
		// or several modules if the modules with revision were found (see https://github.com/openconfig/goyang/issues/137)
		// this loop gets the name of the module
		// and removing the modules with revisions, since the bare module refers to the same module
		for k, m := range ms.Modules {
			if strings.Contains(k, "@") {
				delete(ms.Modules, k) // removing module with a revision
				continue
			}
			mn = m.Name
		}
		// fmt.Println(ms.Modules)
		// os.Exit(0)
		errs := ms.Process()
		for _, err := range errs {
			log.Fatal(err)
		}

		e := yang.ToEntry(ms.Modules[mn])

		var paths []*path.Path
		path.Paths(e, path.Path{}, &paths, !viper.GetBool("path-no-color"))

		// outputting paths in text format
		if viper.GetString("path-format") == "text" {
			for _, path := range paths {

				switch viper.GetString("path-only-nodes") {
				case "config":
					if path.Config == yang.TSFalse {
						continue
					}
				case "state":
					if path.Config == yang.TSTrue || path.Config == yang.TSUnset {
						continue
					}
				}

				var ps []string // path string to print as a slice

				if viper.GetString("path-with-module") == "yes" {
					ps = append(ps, path.Module)
				}

				if viper.GetBool("path-node-state") {
					cfgElem := "[rw]"
					if path.Config == yang.TSFalse {
						cfgElem = "[ro]"
					}
					ps = append(ps, cfgElem)
				}

				switch viper.GetString("path-style") {
				case "xpath":
					ps = append(ps, path.XPath)
				case "restconf":
					ps = append(ps, path.RestConfPath)
				}

				switch viper.GetString("path-types") {
				case "yes":
					ps = append(ps, path.Type.Name)
				case "detailed":
					ps = append(ps, path.SType)
				}
				fmt.Println(strings.Join(ps, "  "))
			}
		}

		if viper.GetString("path-format") == "html" {
			t := viper.GetString("path-template") // path to the template file
			vars := viper.GetStringSlice("path-template-vars")
			if err := path.Template(t, paths, vars); err != nil {
				log.Fatal(err)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringSliceP("yang-dir", "y", []string{"./"}, "directory(-ies) with YANG modules. Values can be provided multiple times (-y dir1 -y dir2)")
	viper.BindPFlag("yang-dir", exportCmd.Flags().Lookup("yang-dir"))

	exportCmd.Flags().StringP("module", "m", "", "path to the YANG file to use for path export (required)")
	viper.BindPFlag("module", exportCmd.Flags().Lookup("module"))
	exportCmd.MarkFlagRequired("module")

	exportCmd.Flags().StringP("format", "f", "text", "paths output format. One of [text, html]")
	viper.BindPFlag("path-format", exportCmd.Flags().Lookup("format"))

	exportCmd.Flags().StringP("style", "s", "xpath", "style of the path. One of [xpath, restconf]")
	viper.BindPFlag("path-style", exportCmd.Flags().Lookup("style"))

	exportCmd.Flags().StringP("with-module", "", "no", "print module name")
	viper.BindPFlag("path-with-module", exportCmd.Flags().Lookup("with-module"))

	exportCmd.Flags().BoolP("node-state", "", true, "print node state")
	viper.BindPFlag("path-node-state", exportCmd.Flags().Lookup("node-state"))

	exportCmd.Flags().StringP("only-nodes", "o", "all", "display only nodes of the given type; one of [all, config, state]")
	viper.BindPFlag("path-only-nodes", exportCmd.Flags().Lookup("only-nodes"))

	exportCmd.Flags().StringP("types", "", "detailed", "display path type information; one of [yes, no, detailed]")
	viper.BindPFlag("path-types", exportCmd.Flags().Lookup("types"))

	exportCmd.Flags().StringP("template", "", "", "path to HTML template to use instead of the default one")
	viper.BindPFlag("path-template", exportCmd.Flags().Lookup("template"))

	exportCmd.Flags().StringSliceP("template-vars", "", []string{}, "extra template variables in case a custom template is used. Key value pairs separated with ::: delimiter")
	viper.BindPFlag("path-template-vars", exportCmd.Flags().Lookup("template-vars"))

	exportCmd.Flags().BoolP("no-color", "", false, "disable colored terminal output")
	viper.BindPFlag("path-no-color", exportCmd.Flags().Lookup("no-color"))
}
