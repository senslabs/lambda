package main

import (
	"net/url"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.CORS())
	url, err := url.Parse("http://router.fission.svc.cluster.local")
	if err != nil {
		e.Logger.Fatal(err)
	}
	targets := []*middleware.ProxyTarget{{URL: url}}
	e.Use(middleware.Proxy(middleware.NewRoundRobinBalancer(targets)))
	e.Logger.Fatal(e.Start(":9806"))
}
