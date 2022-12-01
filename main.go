/*
// Implementation of a main function setting a few characteristics of
// the game window, creating a game, and launching it
*/

package main

import (
	"flag"
	_ "image/png"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800 // Width of the game window (in pixels)
	screenHeight = 160 // Height of the game window (in pixels)
)


func main() {
	var getTPS bool
	flag.BoolVar(&getTPS, "tps", true, "Afficher le nombre d'appel à Update par seconde")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("BUT2 année 2022-2023, R3.05 Programmation système")

	g := InitGame()

	if g.conn != nil {
		g.getTPS = getTPS

		// on appelle la fonction qui écoute les messages envoyés par le serveur
		// on utilise une goroutine pour ne pas bloquer le reste du main
		go ReadFromServer(&g)

		err := ebiten.RunGame(&g)

		if err != nil {
			log.Print(err)
		}
	}
}