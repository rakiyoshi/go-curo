package main

import "flag"

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "ch1", "set run router mode")
	flag.Parse()

	switch mode {
	case "ch1":
		runChapter1()
	case "ch2":
		runChapter2()
	default:
	}
}
