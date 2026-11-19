package app_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/arquivei/go-app"
)

func ExampleProbeGroup_ServeHTTP() {
	readiness := app.NewProbeGroup("readiness")
	database := readiness.MustNewProbe("database", false)
	readiness.MustNewProbe("cache", true)

	request := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/ready", nil)
	response := httptest.NewRecorder()
	readiness.ServeHTTP(response, request)
	fmt.Println(response.Code, response.Body.String())

	database.SetOk()
	response = httptest.NewRecorder()
	readiness.ServeHTTP(response, request)
	fmt.Println(response.Code, response.Body.String())

	// Output:
	// 500 readiness:database
	// 200 readiness:OK
}

func ExampleNewShutdownFunc() {
	handler := &app.ShutdownHandler{
		Name: "flush",
		Handler: app.NewShutdownFunc(func() {
			fmt.Println("flushed pending work")
		}),
	}

	if err := handler.Execute(context.Background()); err != nil {
		fmt.Println(err)
	}

	// Output:
	// flushed pending work
}
