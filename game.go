/*
//  Data structure for representing a game. Implements the ebiten.Game
//  interface (Update in game-update.go, Draw in game-draw.go, Layout
//  in game-layout.go). Provided with a few utilitary functions:
//    - initGame
*/

package main

import (
	"bufio"
	"bytes"
	"course/assets"
	"image"
	"log"
	"net"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	state                     int           // Current state of the game
	runnerImage               *ebiten.Image // Image with all the sprites of the runners
	runners                   [4]Runner     // The four runners used in the game
	f                         Field         // The running field
	launchStep                int           // Current step in StateLaunchRun state
	resultStep                int           // Current step in StateResult state
	getTPS                    bool          // Help for debug
	conn                      net.Conn 		// la connexion vers le serveur
	id                        int 			// id unique qui permet d'identifier le client
	writer                    *bufio.Writer // le writer qui permet d'écrire sur le serveur
	c                         chan msgContentType // le canal permet de passer les messages du serveur aux parties du code concernées
	nbJoueurs                 string 		// permet de connaître le nombre de joueurs actuellement connectés
	nbOfPlayersReadyToRestart string 		// le nombre de joueurs qui sont prêts à rejouer après la fin de la partie
	isPlayerReadyToRestart    bool 			// permet de savoir si le joueur est prêt à redémarrer
}

// These constants define the five possible states of the game
const (
	StateWelcomeScreen int = iota // Title screen
	StateChooseRunner             // Player selection screen
	StateLaunchRun                // Countdown before a run
	StateRun                      // Run
	StateResult                   // Results announcement
)

// InitGame builds a new game ready for being run by ebiten  
func InitGame() (g Game) {

	// on définit l'ip du serveur, ainsi que le port sur lequel le serveur écoute
	var serverIp string = "172.21.65.60"
	var serverPort string = "8080"

	// Dial the server to join the game
	// on se connecte au serveur en utilisant l'ip et port précédement définis
	conn, err := net.Dial("tcp", serverIp+":"+serverPort)

	if err != nil {
		log.Println("Nous n'avons pas pu vous connecter au serveur")
		log.Println("Dial error:", err)
		return g
	}

	// on sauvegarde la connexion dans g
	g.conn = conn

	// on crée le writer et le canal
	g.writer = bufio.NewWriter(conn)
	g.c = make(chan msgContentType, 1)

	// defer conn.Close()

	// Open the png image for the runners sprites
	img, _, err := image.Decode(bytes.NewReader(assets.RunnerImage))
	if err != nil {
		log.Fatal(err)
	}
	g.runnerImage = ebiten.NewImageFromImage(img)

	// Define game parameters
	start := 50.0
	finish := float64(screenWidth - 50)
	frameInterval := 20

	// Create the runners
	for i := range g.runners {
		// interval := 0
		// if i == 0 {
		// 	interval = frameInterval
		// }
		g.runners[i] = Runner{
			xpos: start, ypos: 50 + float64(i*20),
			maxFrameInterval: frameInterval,
			colorScheme: 0,
		}
	}

	// Create the field
	g.f = Field{
		xstart:   start,
		xarrival: finish,
		chrono:   time.Now(),
	}

	// on définit le nombre de joueurs et le nombre de joueurs prêts à recommencer à 0
	g.nbJoueurs = "0"
	g.nbOfPlayersReadyToRestart = "0"

	// on définit le joueur pas prêt à recommencer
	g.isPlayerReadyToRestart = false

	return g
}
