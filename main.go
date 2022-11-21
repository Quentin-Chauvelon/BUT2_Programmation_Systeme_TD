/*
// Implementation of a main function setting a few characteristics of
// the game window, creating a game, and launching it
*/

package main

import (
	"bufio"
	"flag"
	_ "image/png"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800 // Width of the game window (in pixels)
	screenHeight = 160 // Height of the game window (in pixels)
)

type msgContentType struct {
	msgType        string
	id             int
	nbConnected    string
	selectedScheme int
	runTime        time.Duration
}

func newMsgContent() *msgContentType {
	var msgContent msgContentType = msgContentType{"", 0, "0", 0, time.Since(time.Now())}

	return &msgContent
}

func ReadFromServer(g *Game) {
	var reader *bufio.Reader
	reader = bufio.NewReader(g.conn)

	for {
		msg, err := reader.ReadString('\n')

		if err != nil {
			log.Println("Erreur : ", err)
			return
		}

		if msg != "" {

			s := strings.Split(msg, "|")
			var msgContent = newMsgContent()

			if s != nil && len(s) > 0 {
				switch s[0] {

				case "id":
					msgContent.msgType = "id"
					msgContent.id, _ = strconv.Atoi(s[1])

				case "waitingForPlayers":
					msgContent.msgType = "waitingForPlayers"
					msgContent.nbConnected = s[1]

				case "playerSelectedRunner":
					msgContent.msgType = "playerSelectedRunner"
					msgContent.selectedScheme, _ = strconv.Atoi(s[1])

				case "startCountdown":
					msgContent.msgType = "startCountdown"

				case "runnerArrived":
					msgContent.msgType = "runnerArrived"
					msgContent.runTime, err = time.ParseDuration(s[1])

					if err != nil {
						log.Println("Erreur : ", err)
						return
					}

				case "showResults":
					msgContent.msgType = "showResults"

				case "playerIsReadyToRestart":
					msgContent.msgType = "playerIsReadyToRestart"
					msgContent.nbConnected = s[1]
				}
			}

			g.c <- *msgContent
		}
	}
}

// func WriteToServer(conn net.Conn, message string) {
// 	var writer *bufio.Writer
// 	writer = bufio.NewWriter(conn)

// 	log.Println("writing to server ", message)
// 	writer.WriteString(message + "|\n")
// 	writer.Flush()
// }

func WriteToServer(writer *bufio.Writer, message string) {
	log.Println("writing to server ", message)
	writer.WriteString(message + "|\n")
	writer.Flush()
}

func main() {

	var getTPS bool
	flag.BoolVar(&getTPS, "tps", true, "Afficher le nombre d'appel à Update par seconde")
	flag.Parse()

	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("BUT2 année 2022-2023, R3.05 Programmation système")

	g := InitGame()

	if g.conn != nil {
		g.getTPS = getTPS

		go ReadFromServer(&g)

		err := ebiten.RunGame(&g)

		if err != nil {
			log.Print(err)
		}
	}
}

// faire un serveur propre (après le accept, goroutine pour chaque client)

// struct sur le serveur pour stocker la couleur du perso et s'il est sélectionné?
// et pour sélectionner la position exacte en temps réel pour sauter par dessus, et c'est le serveur qui dit ou chaque client doit se positionner quand il bouge

// empêcher deux fois la même couleur
// annuler une sélection
// montrer en temps réel où tu es + sauter par dessus quelqu'un

// voir en temps réel les coureurs
// montrer les personnages avec les bonnes couleurs et les bons temps
// avoir le même ordre de runners sur chaque client
