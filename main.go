/*
// Implementation of a main function setting a few characteristics of
// the game window, creating a game, and launching it
*/

package main

import (
	"flag"
	_ "image/png"
	"log"
	"strings"
	"strconv"
	"bufio"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	screenWidth  = 800 // Width of the game window (in pixels)
	screenHeight = 160 // Height of the game window (in pixels)
)

 	

type msgContentType struct {
    msgType 		string
    nbConnected  	string
	selectedScheme	int
}


func newMsgContent() *msgContentType {
	var msgContent msgContentType = msgContentType{"", "0", 0}

	return &msgContent
}



func ReadFromServer(g *Game) {
	var reader *bufio.Reader
	reader = bufio.NewReader(g.conn)

	for  {
		msg, err := reader.ReadString('\n')

		if (err != nil) {
			log.Println("Erreur : ", err)
			return
		}

		if (msg != "") {

			s := strings.Split(msg, "|")
			var msgContent = newMsgContent()

			if (s != nil && len(s) > 0) {
				switch s[0] {

					case "waitingForPlayers":
						msgContent.msgType = "waitingForPlayers"
						msgContent.nbConnected = s[1]

					case "playerSelectedRunner":
						msgContent.msgType = "playerSelectedRunner"
						msgContent.selectedScheme, _ = strconv.Atoi(s[1]) 
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