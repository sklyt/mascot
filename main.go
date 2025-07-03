package main

import mascot "github.com/mascot/maskot"

func main() {
	win := mascot.GetMaskot([]string{
		// https://opengameart.org/content/dancing-girl-sprites
		// "C:/[path]/mascot1.png",
		// "C:/[path]//mascot2.png",
		// "C:/[path]//mascot3.png",,

	}, 8)

	win.Run()
}
