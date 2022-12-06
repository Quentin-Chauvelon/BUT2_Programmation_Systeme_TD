package main

import "time"

// lit dans le canal durant StateRun
// la fonction est appelé plusieurs fois en tant que goroutine, ce qui permet d'attendre un message ou d'arrêter
// le select après un certain temps sans bloquer l'exécution du reste du programme
func getStateRunMessages(g *Game) {
	// on lit dans le canal
	select {
	case msg := <-g.c:

		// on met à jour la position et la vitesse du runner donné
		if msg.msgType == "updateRunnerPosition" {
			g.runners[msg.id].xpos = msg.runnerPosition
			g.runners[msg.id].speed = msg.runnerSpeed

		}

		// un runner est arrivé, on modifie son temps
		if msg.msgType == "runnerArrived" {
			g.runners[msg.id].runTime = msg.runTime
			g.runners[msg.id].arrived = true

			// on affiche les résultats
		}

		// afficher les résultats
		if msg.msgType == "showResults" {
			g.state++
		}

	// au bout de 16 millisecondes (~1 frame à 60fps), on arrête la goroutine, s'il ne s'est rien passé
	case <- time.After(16 * time.Millisecond):
	}
}