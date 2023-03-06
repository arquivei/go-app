package main

import "github.com/arquivei/go-app"

type config struct {
	// App is the app specific configuration
	app.Config

	// Programs can have any configuration the want.

	HTTP struct {
		Port string `default:"8000"`
	}
	Dir string `default:"."`
}
