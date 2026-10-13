package main

import (
	"github.com/spf13/cobra"
	"github.com/ubgo/voltkit/skillcmd"
)

// newSkillsCommand serves this project's published skills — content fetched
// once per binary version into the OS cache and served from there, always
// matching this binary. See github.com/ubgo/voltkit/skillcmd.
func newSkillsCommand() *cobra.Command {
	return skillcmd.New(skillcmd.Options{
		Binary:  "[[.Binary]]",
		Repo:    "[[.Repo]]",
		Version: version, // the project's stamped version variable
		Tag:     "[[.Tag]]" + version,
		// The detected forge's download shape, fixed at gen time — the
		// binary never guesses its host.
		AssetURLTemplate: "[[.AssetTpl]]",
	})
}
