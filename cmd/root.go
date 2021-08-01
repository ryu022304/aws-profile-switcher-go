/*
Copyright © 2021 Ryunosuke Makihara <ryu022304@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Profile is struct for display selector area
type Profile struct {
	Name   string
	Region string
	Output string
}

// Data if struct for profile's data
type Data struct {
	Region string
	Output string
}

// bellSkipper implements an io.WriteCloser that skips the terminal bell
// character (ASCII code 7), and writes the rest to os.Stderr. It is used to
// replace readline.Stdout, that is the package used by promptui to display the
// prompts.
//
// This is a workaround for the bell issue documented in
// https://github.com/manifoldco/promptui/issues/49.
type bellSkipper struct{}

// Write implements an io.WriterCloser over os.Stderr, but it skips the terminal
// bell character.
func (bs *bellSkipper) Write(b []byte) (int, error) {
	const charBell = 7 // c.f. readline.CharBell
	if len(b) == 1 && b[0] == charBell {
		return 0, nil
	}
	return os.Stderr.Write(b)
}

// Close implements an io.WriterCloser over os.Stderr.
func (bs *bellSkipper) Close() error {
	return os.Stderr.Close()
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "aws-ps",
	Short: "AWS CLI profile switcher",
	Long: `You can easily switch AWS CLI profile settings.
     __          _______   _____            __ _ _         _____         _ _       _ 
    /\ \        / / ____| |  __ \          / _(_) |       / ____|       (_) |     | |
   /  \ \  /\  / / (___   | |__) | __ ___ | |_ _| | ___  | (_____      ___| |_ ___| |__   ___ _ __ 
  / /\ \ \/  \/ / \___ \  |  ___/ '__/ _ \|  _| | |/ _ \  \___ \ \ /\ / / | __/ __| '_ \ / _ \ '__|
 / ____ \  /\  /  ____) | | |   | | | (_) | | | | |  __/  ____) \ V  V /| | || (__| | | |  __/ |
/_/    \_\/  \/  |_____/  |_|   |_|  \___/|_| |_|_|\___| |_____/ \_/\_/ |_|\__\___|_| |_|\___|_|
	`,
	Run: func(cmd *cobra.Command, args []string) {
		execCommand()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func execCommand() {
	// Get AWS_PROFILE env value
	currentProfile := os.Getenv("AWS_PROFILE")
	if currentProfile == "" {
		currentProfile = "default"
	}

	// Read ~/.aws/config
	viper.AddConfigPath("$HOME/.aws")
	viper.SetConfigType("ini")
	viper.SetConfigName("config")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	settings := viper.AllSettings()

	// Create prompt selector
	profile := execPromptui(settings)

	// Set AWS_PROFILE env value
	if profile == "default" {
		fmt.Println("AWS_PROFILE=''")
	} else {
		fmt.Println("AWS_PROFILE='" + profile + "'")
	}
	writeProfile(profile)
}

func execPromptui(settings map[string]interface{}) string {
	profiles := []Profile{}
	for key, value := range settings {
		words := strings.Split(key, " ")
		var name string
		if len(words) > 1 {
			name = strings.Join(words[1:], " ") // profile hoge fuga
		} else {
			name = words[0] // default
		}
		var temp Data
		jsonStr, err := json.Marshal(value)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err := json.Unmarshal(jsonStr, &temp); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		profiles = append(profiles, Profile{Name: name, Region: temp.Region, Output: temp.Output})
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}?",
		Active:   "\U000027A1 {{ .Name | yellow }} ({{ .Region | red }})",
		Inactive: "  {{ .Name | cyan }} ({{ .Region | red }})",
		Selected: "\U000027A1 {{ .Name | red | cyan }}",
		Details: `
--------- Profile ----------
{{ "Name  :" | faint }}  {{ .Name }}
{{ "Region:" | faint }}  {{ .Region }}
{{ "Output:" | faint }}  {{ .Output }}`,
	}

	prompt := promptui.Select{
		Label:     "AWS CLI profile",
		Items:     profiles,
		Templates: templates,
		Size:      5,
		Stdout:    &bellSkipper{},
	}

	i, _, err := prompt.Run()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return profiles[i].Name
}

func writeProfile(profile string) {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	configPath := filepath.Join(home, ".aws-ps")
	f, err := os.Create(configPath)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()
	f.WriteString(profile)
}
