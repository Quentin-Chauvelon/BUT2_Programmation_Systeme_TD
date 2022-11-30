/*
//  Implementation of the Update method for the Game structure
//  This method is called once at every frame (60 frames per second)
//  by ebiten, juste before calling the Draw method (game-draw.go).
//  Provided with a few utilitary methods:
//    - CheckArrival
//    - ChooseRunners
//    - HandleLaunchRun
//    - HandleResults
//    - HandleWelcomeScreen
//    - Reset
//    - UpdateAnimation
//    - UpdateRunners
*/

package main

import (
	"strconv"
	"time"
	"fmt"
	// "log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HandleWelcomeScreen waits for the player to push SPACE in order to
// start the game
// la fonction lit aussi dans le canal pour définir l'id du client,
// pour changer le nombre de joueurs connectés
// et pour modifier le runner sélectionné et/ou validé par un client
func (g *Game) HandleWelcomeScreen() bool {

	// on lit dans le canal
	select {
	case msg := <-g.c:

		// permet de définir l'id du joueur
		if msg.msgType == "id" {
			g.id = msg.id

			// on modifie la couleur du runner sélectionnée pour qu'elle corresponde à l'id
			// cela permet d'avoir une couleur unique pour chaque runner dès le début de la sélection des runners
			g.runners[g.id].colorScheme = g.id + 1

		// permet de savoir le nombre de joueurs connectés
		} else if msg.msgType == "waitingForPlayers" {
			g.nbJoueurs = msg.nbConnected

			if g.nbJoueurs == "4" {
				go WriteToServer(g.writer, "playerChangedRunner|right")
			}

		// sélectionne le runner donné pour le joueur donné
		} else if msg.msgType == "playerChangedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme

		// valide ou annule la sélection du runner donné pour le joueur donné
		} else if msg.msgType == "playerSelectedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme
			g.runners[msg.id].colorSelected = !g.runners[msg.id].colorSelected
		}

	default:
	}

	// on passe à l'étape suivante si le joueur appuie sur espace et qu'il y a 4 joueurs connectés
	return inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.nbJoueurs == "4"
}


// ChooseRunners permet de sélectionner un runner (lorsque l'on appuie sur les flèches gauche ou droite)
//
func (g *Game) ChooseRunners() {

	// 	done = true

	// 	done = g.runners[0].ManualChoose() && done

	// 	if done {
		// 		go WriteToServer(g.writer, "playerSelectedRunner|" + strconv.Itoa(g.runners[0].colorScheme))
	// 	}

	// 	return done

	// si le joueur n'a pas encore choisi de runner
	if !(g.runners[g.id].colorSelected) {

		// s'il valide sa sélection, on envoie au serveur le runner sélectionné
		if g.runners[g.id].ManualChoose() {
			go WriteToServer(g.writer, "playerSelectedRunner|" + strconv.Itoa(g.runners[g.id].colorScheme))
		}

		// s'il se déplace à gauche ou droite, on l'envoie au serveur pour que celui-ci nous dise
		// quel devrait maintenant être sélectionné
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
			go WriteToServer(g.writer, "playerChangedRunner|right")
		} else if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
			go WriteToServer(g.writer, "playerChangedRunner|left")
		}

	// Si le joueur à déjà sélectionné un runner et qu'il appuie sur espace, alors cela annule sa sélection
	} else {
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			go WriteToServer(g.writer, "playerSelectedRunner|")
		}
	}

	// on lit dans le canal
	select {
	case msg := <-g.c:

		// sélectionne le runner donné pour le joueur donné
		if msg.msgType == "playerChangedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme

		// valide ou annule la sélection du runner donné pour le joueur donné
		}  else if msg.msgType == "playerSelectedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme
			g.runners[msg.id].colorSelected = !g.runners[msg.id].colorSelected

		// démarre le compte à rebours
		} else if msg.msgType == "startCountdown" {
			g.UpdateAnimation()
			g.state++
		}

	default:
	}
}

// HandleLaunchRun countdowns to the start of a run
func (g *Game) HandleLaunchRun() bool {
	if time.Since(g.f.chrono).Milliseconds() > 1000 {
		g.launchStep++
		g.f.chrono = time.Now()
	}
	if g.launchStep >= 5 {
		g.launchStep = 0
		return true
	}
	return false
}


func lirePosition(g *Game) {
	// on lit dans le canal
	select {
	case msg := <-g.c:

		// on met à jour la position et la vitesse du runner donné
		if msg.msgType == "updateRunnerPosition" {
			g.runners[msg.id].xpos = msg.runnerPosition
			g.runners[msg.id].speed = msg.runnerSpeed
		}

	case <- time.After(16 * time.Millisecond):
	}
}


// UpdateRunners
func (g *Game) UpdateRunners() {

	// for i := range g.runners {
		// if i == 0 {
				// g.runners[i].ManualUpdate()
			// } else {
					// g.runners[i].RandomUpdate()
			// }
		// }
	// }

	// si le runner n'est pas encore arrivé
	if !g.runners[g.id].arrived {
		// on sauvegarde sa position
		var previousPosition float64 = g.runners[g.id].xpos

		// on met à jour sa position
		g.runners[g.id].ManualUpdate()

		// s'il le runner a changé de position, on envoie sa nouvelle position au serveur
		// (cela évite de le faire systématiquement même si le joueur ne bouge pas)
		if g.runners[g.id].xpos != previousPosition {
			go WriteToServer(g.writer, "updateRunnerPosition|" + fmt.Sprintf("%f", g.runners[g.id].xpos) + "|" + fmt.Sprintf("%f", g.runners[g.id].speed))
		}
	}

	// on lit dans le canal
	select {
	case msg := <-g.c:

		// on met à jour le runTime du runner donné
		if msg.msgType == "runnerArrived" {
			g.runners[msg.id].runTime = msg.runTime
			g.runners[msg.id].arrived = true

			// on affiche les résultats
		} else if msg.msgType == "showResults" {
			g.state++
		}

	default:
	}

	for i := 0; i < 4; i++ {
		go lirePosition(g)
	}
}


// CheckArrival regarde si le runner est arrivé, et le dit au serveur s'il l'est
func (g *Game) CheckArrival() {

	// finished = true
	// for i := range g.runners {
	// 	g.runners[i].CheckArrival(&g.f)
	// 	finished = finished && g.runners[i].arrived
	// }
	// return finished

	// on regarde si le runner est arrivé
	g.runners[g.id].CheckArrival(&g.f)

	// si le runner est arrivé, on envoie au serveur notre temps
	if g.runners[g.id].arrived {
		go WriteToServer(g.writer, "runnerArrived|"+g.runners[g.id].runTime.String())
	}
}

// Reset resets all the runners and the field in order to start a new run
func (g *Game) Reset() {
	for i := range g.runners {
		g.runners[i].Reset(&g.f)
	}
	g.f.Reset()
}

// UpdateAnimation loops over all the runners to update their sprite
func (g *Game) UpdateAnimation() {
	for i := range g.runners {
		g.runners[i].UpdateAnimation(g.runnerImage)
	}
}

// HandleResults computes the resuls of a run and prepare them for
// being displayed
// la fonction 
func (g *Game) HandleResults() {

	if !g.isPlayerReadyToRestart {
		if time.Since(g.f.chrono).Milliseconds() > 1000 || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.resultStep++
			g.f.chrono = time.Now()
		}

		if g.resultStep >= 4 && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			// g.resultStep = 0
			// return true

			// on définit le joueur comme prêt à redémarrer et on le dit au serveur
			g.isPlayerReadyToRestart = true
			go WriteToServer(g.writer, "playerIsReadyToRestart|")
		}
	}


	// on lit dans le canal
	select {
	case msg := <-g.c:

		// un joueur est prêt à redémarrer
		if msg.msgType == "playerIsReadyToRestart" {

			// on modifie le nombre de joueurs prêts à redémarrer
			g.nbOfPlayersReadyToRestart = msg.nbConnected

			// si les 4 joueurs sont prêts à redémarrer, on réinitialise certaines variables de g et on recommence le compte à rebours
			if msg.nbConnected == "4" {
				g.Reset()
				g.state = StateLaunchRun

				g.resultStep = 0

				g.isPlayerReadyToRestart = false
				g.nbOfPlayersReadyToRestart = "0"
			}
		}

	default:
	}
}

// Update is the main update function of the game. It is called by ebiten
// at each frame (60 times per second) just before calling Draw (game-draw.go)
// Depending of the current state of the game it calls the above utilitary
// function and then it may update the state of the game
func (g *Game) Update() error {
	switch g.state {
	case StateWelcomeScreen:
		done := g.HandleWelcomeScreen()
		if done {
			g.state++
		}
	// case StateChooseRunner:
	// 	done := g.ChooseRunners()
	// 	if done {
	// 		g.UpdateAnimation()
	// 		g.state++
	// 	}

	case StateChooseRunner:
		// done := g.ChooseRunners()
		// if done {
		// 	g.UpdateAnimation()
		// 	g.state++
		// }
		g.ChooseRunners()

	case StateLaunchRun:
		done := g.HandleLaunchRun()
		if done {
			g.state++
		}

	case StateRun:
		// g.UpdateRunners()
		// finished := g.CheckArrival()
		// g.UpdateAnimation()
		// if finished {
			// g.state++
		// }

		// si le runner n'est pas encore arrivé, on regarde s'il le sera après avoir mis à jour sa position
		if !g.runners[g.id].arrived {
			g.CheckArrival()
		}

		g.UpdateRunners()

		g.UpdateAnimation()

	case StateResult:
		// done := g.HandleResults()
		// if done {
		// 	g.Reset()
		// 	g.state = StateLaunchRun
		// }
		g.HandleResults()
	}

	return nil
}
