package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/app"
	"github.com/tortoise10h/calculus/internal/workspace"
)

func main() {
	// SetName before the first Path call, so PathUserData resolves to a stable
	// …/Application Support/Calculus instead of whatever the bundle reports.
	mygo.App.SetName("Calculus")

	calcApp := loadWorkspace()

	mygo.Theme.SetSource(mygo.ThemeLight)

	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	}))

	// Save synchronously on quit, so the last keystrokes before Cmd+Q survive.
	mygo.App.OnBeforeQuit(func(*mygo.QuitEvent) { calcApp.Flush() })

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:                "Calculus",
			Width:                720,
			Height:               520,
			MinWidth:             480,
			MinHeight:            360,
			TitleBarStyle:        mygo.TitleBarHiddenInset,
			TrafficLightPosition: &mygo.Point{X: 18, Y: 13},
			BackgroundColor:      "#ffffff",
			StateKey:             "main",
			Content:              ui.View(calcApp.View),
		})
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// loadWorkspace restores the saved tabs, or seeds a starter tab on first run.
// Every failure path here is non-fatal: a missing, unreadable or corrupt
// workspace still yields a fully usable app, just one that will not remember
// anything yet.
func loadWorkspace() *app.App {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		log.Printf("calculus: user data dir unavailable, running without auto-save: %v", err)
		return app.New()
	}

	store := workspace.NewFileStore(dir)
	ws, err := workspace.LoadOrSeed(store)
	if err != nil {
		log.Printf("calculus: workspace: %v", err)
	}
	return app.NewWithWorkspace(ws, store)
}
