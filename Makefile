
install:
	go install github.com/crackcomm/actions-cli/app-build

example: install
	app-build -app example/app.yaml -o exampleapp -name exampleapp
