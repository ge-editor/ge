package main

import (
	"context"
	"fmt"

	"github.com/ge-editor/editorleaf"
	"github.com/ge-editor/gecore/excommand"
	"github.com/ge-editor/gecore/killbuffer"
)

// Implements

type Command struct {
	CommandName        string
	CommandAliases     []string
	CommandDescription string
	Run                func(ctx context.Context, args []string) (handled bool, err error)
}

func (c *Command) Name() string        { return c.CommandName }
func (c *Command) Aliases() []string   { return c.CommandAliases }
func (c *Command) Description() string { return c.CommandDescription }

func (c *Command) Execute(ctx context.Context, args []string) (bool, error) {
	if c.Run == nil {
		return false, nil
	}
	return c.Run(ctx, args)
}

func init() {

	excommand.RegisterGlobal(&Command{
		CommandName:        "info",
		CommandDescription: "Set editor information in kill-buffer.",
		Run: func(ctx context.Context, args []string) (bool, error) {
			l := len(*editorleaf.BufferSets)
			s := ver() + "\n\n"
			s += "editorleaf.BufferSets:\n"
			for i := 0; i < l; i++ {
				bs := (*editorleaf.BufferSets)[i]
				s += fmt.Sprintf("- %02d: %s (Metas: %d)\n", i, bs.GetPath(), len(bs.GetMetas()))
			}
			s += "\n"

			// Convert string to [][]byte
			r, _ := editorleaf.SplitRows([]byte(s))
			killbuffer.KillBuffer.PushKillBuffer(r.BytesArray(), []byte{'\n'})

			return true, nil
		},
	})

}
