package main

func main() {
	router := NewRouter()
	err := router.Run(":1024")
	if err != nil {
		return
	}
}
