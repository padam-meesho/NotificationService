package main

import (
	"github.com/padam-meesho/NotificationService/config"
	"github.com/padam-meesho/NotificationService/internal/app"
	"github.com/padam-meesho/NotificationService/internal/models"
	"github.com/padam-meesho/NotificationService/internal/routes"
)

type appConfig struct {
	Configs models.AppConfig
}

var (
	appConfigInstance appConfig
)

func main() {
	// all the server initialization and other logic shall be placed here.
	// other important key dependencies shall be placed here too.
	// have a function that sets up the global configs from the ENV.
	// put all of these setup functions inside a single newapp function.
	config.LoadAppConfig(&appConfigInstance.Configs)
	app.NewApp(appConfigInstance.Configs)
	routes.SetUpRoutes()
}

// additionally for all the different configs or the services,
// try to have an interface and then an implementation of that interface.
