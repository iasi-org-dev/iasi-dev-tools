package main

import (
	"iasi-dev/internal/runners"
	"iasi-dev/internal/structures"
)

type CommandHandler func(*structures.Context)

type Command struct {
	Run CommandHandler
}

var Registry = map[string]Command{
	"build":         {Run: func(c *structures.Context) { runners.Build(c, 0) }},
	"publish":       {Run: func(c *structures.Context) { runners.Publish(c, 0) }},
	"commit":        {Run: func(c *structures.Context) { runners.Commit(c, 0) }},
	"release":       {Run: func(c *structures.Context) { runners.Release(c, 0) }},
	"workflow":      {Run: runners.Workflow},
	"sync":          {Run: runners.Sync},
	"version":       {Run: runners.Version},
	"freeze":        {Run: func(c *structures.Context) { runners.Freeze(c) }},
	"promote":       {Run: func(c *structures.Context) { runners.Promote(c) }},
	"promote-check": {Run: func(c *structures.Context) { runners.PromoteCheck(c) }},
	"restore":       {Run: func(c *structures.Context) { runners.Restore(c) }},
	"materialize":   {Run: func(c *structures.Context) { runners.Materialize(c) }},
}
