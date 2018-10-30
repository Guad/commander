# Commander

Commander is a simple chat command parser.

### Simple Usage

````
func main() {
	c := commander.New()
	c.Use(logMiddleware)

	c.Command("/hello {times:int} {text+}", helloCommand)

	text := "/hello 5 world and everyone"
	ok, err := c.Execute(strings.Trim(text, "\n\r"))
	fmt.Println("Success?", ok, "Error:", err)
}
````
