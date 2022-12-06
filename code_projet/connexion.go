/*
Fichier contenant les fonctions permettant la connexion
et la communication entre le serveur et le client
*/

package main

import (
	"bufio"
	"log"
	"strconv"
	"strings"
	"time"
)

type msgContentType struct {
	msgType        string        // le type du message (permet de savoir à quelles parties du code le message est destiné)
	id             int           // l'id du client qui a envoyé le message au serveur
	nbConnected    string        // le nombre de joueurs connectés au moment où le message a été envoyé au serveur
	selectedScheme int           // la couleur du runner sélectionné par le joueur qui a envoyé le message au serveur
	runTime        time.Duration // le temps final du joueur qui a envoyé le message au serveur
	runnerPosition float64       // la position du runner du joueur qui l'a envoyée au serveur
	runnerSpeed    float64       // la vitesse du runner du joueur qui l'a envoyée au serveur
}

// ReadFromServer est une boucle infinie qui reçoit les messages envoyés par le serveur
func ReadFromServer(g *Game) {
	// comme cette fonction n'est appelé qu'une fois, on peut créer le reader ici et ce sera le même qui sera utilisé tout au long de l'exécution du programme
	var reader *bufio.Reader
	reader = bufio.NewReader(g.conn)

	for {
		// on lit les messages envoyés par le serveur
		msg, err := reader.ReadString('\n')

		if err != nil {
			log.Println("Erreur : ", err)
			return
		}

		// si on a reçu un message
		if msg != "" {

			// on découpe le string qui est sous la forme commence|argument1|argument2|...|\n
			s := strings.Split(msg, "|")
			var msgContent = msgContentType{nbConnected: "0", runTime: time.Since(time.Now())}

			// si valide
			if s != nil && len(s) > 0 {

				// la première case du tableau nous permet de savoir où nous sommes rendus dans le déroulement du jeu
				// et ce que nous sommes ainsi censés faire
				switch s[0] {

				// permet de définir l'id du client durant StateWelcomeScreen
				case "id":
					msgContent.msgType = "id"
					msgContent.id, _ = strconv.Atoi(s[1])

				// permet de connaître le nombre de joueurs connectés durant StateWelcomeScreen
				case "waitingForPlayers":
					msgContent.msgType = "waitingForPlayers"
					msgContent.nbConnected = s[1]

				// permet de modifier la sélection de runner d'un client
				// (quand il s'est déplacé avec les flèches) durant StateWelcomeScreen et StateChooseRunner
				case "playerChangedRunner":
					msgContent.msgType = "playerChangedRunner"
					msgContent.id, _ = strconv.Atoi(s[1])
					msgContent.selectedScheme, _ = strconv.Atoi(s[2])

				// permet de valider ou d'annuler la sélection de runner d'un client durant StateWelcomeScreen et StateChooseRunner
				case "playerSelectedRunner":
					msgContent.msgType = "playerSelectedRunner"
					msgContent.id, _ = strconv.Atoi(s[1])
					msgContent.selectedScheme, _ = strconv.Atoi(s[2])

				// permet de démarrer le compte à rebours durant StateChooseRunner
				case "startCountdown":
					msgContent.msgType = "startCountdown"

				// permet de mettre à jour la position et la vitesse du runner d'un client durant StateRun
				case "updateRunnerPosition":
					msgContent.msgType = "updateRunnerPosition"
					msgContent.id, _ = strconv.Atoi(s[1])
					msgContent.runnerPosition, _ = strconv.ParseFloat(s[2], 64)
					msgContent.runnerSpeed, _ = strconv.ParseFloat(s[3], 64)

				// permet de savoir quand le runner d'un client est arrivé durant StateRun
				case "runnerArrived":
					msgContent.msgType = "runnerArrived"
					msgContent.id, _ = strconv.Atoi(s[1])
					msgContent.runTime, err = time.ParseDuration(s[2])

					if err != nil {
						log.Println("Erreur : ", err)
						return
					}

				// affiche les résultats durant StateRun
				case "showResults":
					msgContent.msgType = "showResults"

				// permet de savoir si le joueur a appuyé sur espace pour redémarrer durant StateResult
				case "playerIsReadyToRestart":
					msgContent.msgType = "playerIsReadyToRestart"
					msgContent.nbConnected = s[1]
				}
			}

			// on ajoute le message dans le canal pour qu'il soit lu par la partie du code concernée
			g.c <- msgContent
		}
	}
}

// WriteToServer envoie au serveur le message passé en paramètre
func WriteToServer(writer *bufio.Writer, message string) {
	// log.Println("writing to server ", message)
	_, err := writer.WriteString(message + "|\n")
	if err != nil {
		log.Println("Erreur de transmission : transformation en string", err)
	}
	err = writer.Flush()
	if err != nil {
		log.Println("Erreur de transmission : nettoyage de buffer", err)
	}
}
