package main

import "github.com/weplanx/transfer/bootstrap"

func main() {
	values, err := bootstrap.SetValues()
	if err != nil {
		panic(err)
	}
	app, err := App(values)
	if err != nil {
		panic(err)
	}
	app.Serve("tcp", values.Address)
}