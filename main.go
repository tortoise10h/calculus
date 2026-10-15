package main

import (
	"log"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/tortoise10h/calculus/internal/app"
)

func main() {
	calcApp := app.New()

	mygo.Theme.SetSource(mygo.ThemeLight)

	mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
		{Role: mygo.RoleAppMenu},
		{Role: mygo.RoleEditMenu},
		{Role: mygo.RoleViewMenu},
		{Role: mygo.RoleWindowMenu},
	}))

	mygo.App.WhenReady(func() {
		mygo.NewWindow(mygo.WindowOptions{
			Title:           "Calculus",
			Width:           720,
			Height:          520,
			MinWidth:        480,
			MinHeight:       360,
			TitleBarStyle:   mygo.TitleBarHiddenInset,
			BackgroundColor: "#ffffff",
			StateKey:        "main",
			Content:         ui.View(calcApp.View),
		})
	})

	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}
