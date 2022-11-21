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
func (g *Game) HandleWelcomeScreen() bool {
	select {
	case msg := <-g.c:

		if msg.msgType == "id" {
			g.id = msg.id

		} else if msg.msgType == "waitingForPlayers" {
			g.nbJoueurs = msg.nbConnected

		} else if msg.msgType == "playerChangedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme

		} else if msg.msgType == "playerSelectedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme
			g.runners[msg.id].colorSelected = !g.runners[msg.id].colorSelected
		}

	default:
	}

	return inpututil.IsKeyJustPressed(ebiten.KeySpace) && g.nbJoueurs == "4"
}

// ChooseRunners loops over all the runners to check which sprite each
// of them selected
// func (g *Game) ChooseRunners() (done bool) {
// 	done = true

// 	done = g.runners[0].ManualChoose() && done

// 	if done {
// 		go WriteToServer(g.writer, "playerSelectedRunner|" + strconv.Itoa(g.runners[0].colorScheme))
// 	}

// 	return done
// }

func (g *Game) ChooseRunners() {

	// si le joueur n'a pas encore choisi de runner, on attend qu'il en choisisse un
	if !(g.runners[g.id].colorSelected) {
		if g.runners[g.id].ManualChoose() {
			go WriteToServer(g.writer, "playerSelectedRunner|" + strconv.Itoa(g.runners[g.id].colorScheme))
		}

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

	select {
	case msg := <-g.c:

		if msg.msgType == "playerChangedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme

		}  else if msg.msgType == "playerSelectedRunner" {
			g.runners[msg.id].colorScheme = msg.selectedScheme
			g.runners[msg.id].colorSelected = !g.runners[msg.id].colorSelected

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

// UpdateRunners loops over all the runners to update each of them
// func (g *Game) UpdateRunners() {
// 	for i := range g.runners {
// 		if i == 0 {
// 			g.runners[i].ManualUpdate()
// 		} else {
// 			g.runners[i].RandomUpdate()
// 		}
// 	}
// }

// UpdateRunners loops over all the runners to update each of them
func (g *Game) UpdateRunners() {

	if !g.runners[g.id].arrived {
		var previousPosition float64 = g.runners[g.id].xpos

		g.runners[g.id].ManualUpdate()

		if g.runners[g.id].xpos != previousPosition {
			go WriteToServer(g.writer, "updateRunnerPosition|" + fmt.Sprintf("%f", g.runners[g.id].xpos) + "|" + fmt.Sprintf("%f", g.runners[g.id].speed))
		}
	}

	select {
	case msg := <-g.c:
		if msg.msgType == "updateRunnerPosition" {
			g.runners[msg.id].xpos = msg.runnerPosition
			g.runners[msg.id].speed = msg.runnerSpeed


		} else if msg.msgType == "runnerArrived" {

			g.runners[msg.id].runTime = msg.runTime
			g.runners[msg.id].arrived = true
			// for i := range g.runners {

			// 	// On modifie le temps du premier joueur qui n'est pas encore arrivé
			// 	if i != 0 && !g.runners[i].arrived {
			// 		g.runners[i].runTime = msg.runTime
			// 		g.runners[i].arrived = true

			// 		// Si tous les joueurs ont fini la course, on montre le résultat
			// 		// (ne fonctionne pas car on passe à l'état suivant avant que la fonction checkArrival soit appelé,
			// 		// ce qui empêche de prévenir le serveur et donc les autres joueurs)
			// 		// if i == 3 {
			// 		// 	g.state++
			// 		// }

			// 		break
			// 	}
			// }

		} else if msg.msgType == "showResults" {
			g.state++
		}

	default:
	}
}

// CheckArrival loops over all the runners to check which ones are arrived
// func (g *Game) CheckArrival() (finished bool) {
// 	finished = true
// 	for i := range g.runners {
// 		g.runners[i].CheckArrival(&g.f)
// 		finished = finished && g.runners[i].arrived
// 	}

// 	rPressed := false

// 	if (inpututil.IsKeyJustPressed(ebiten.KeyR)) {
// 		rPressed = true
// 	}

// 	return finished || rPressed
// }

// CheckArrival loops over all the runners to check which ones are arrived
func (g *Game) CheckArrival() {

	g.runners[g.id].CheckArrival(&g.f)

	if g.runners[g.id].arrived {
		// go WriteToServer(g.writer, "runnerArrived|" + strconv.FormatInt(g.runners[0].runTime.Milliseconds(), 10))
		fmt.Println("runner arrived", g.id)
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
func (g *Game) HandleResults() {

	if !g.isPlayerReadyToRestart {
		if time.Since(g.f.chrono).Milliseconds() > 1000 || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			g.resultStep++
			g.f.chrono = time.Now()
		}

		if g.resultStep >= 4 && inpututil.IsKeyJustPressed(ebiten.KeySpace) {

			g.isPlayerReadyToRestart = true

			go WriteToServer(g.writer, "playerIsReadyToRestart|")
		}
	}

	select {
	case msg := <-g.c:
		if msg.msgType == "playerIsReadyToRestart" {

			g.nbOfPlayersReadyToRestart = msg.nbConnected

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
		g.ChooseRunners()

	case StateLaunchRun:
		done := g.HandleLaunchRun()
		if done {
			g.state++
		}

	case StateRun:
		if !g.runners[g.id].arrived {
			g.CheckArrival()
		}

		g.UpdateRunners()

		g.UpdateAnimation()

	// case StateResult:
	// 	done := g.HandleResults()
	// 	if done {
	// 		g.Reset()
	// 		g.state = StateLaunchRun
	// 	}
	// }

	case StateResult:
		g.HandleResults()
	}

	return nil
}
