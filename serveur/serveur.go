package main

import (
	"bufio"
	"log"
	"net"
	"strconv"
	"strings"
)

type Client struct {
	id 	  			int 			// l'id qui permet d'identifier le client
	conn   			net.Conn 		// la connexion vers le client
	writer 			*bufio.Writer 	// le writer qui permet d'envoyer des messages au client
	colorScheme     int 			// la couleur du runner sélectionné par le client
	colorSelected	bool 			// si le client a validé ou non sa sélection de couleur du runner
}

type Serveur struct {
	clients                []Client // le tableau contenant tous les clients
	nbOfSelectedRunners       int 	// le nombre de clients qui ont validé la sélection de la couleur de leur runner
	nbOfArrivedRunners        int 	// le nombre de runners qui sont arrivés
	nbOfPlayersReadyToRestart int 	// le nombre de clients qui sont prêts à redémarrer
}


// on initialise le serveur avec des valeurs par défaut
var s Serveur = Serveur{[]Client{}, 0, 0, 0}


// lorsqu'un client appuie sur les flèces gauche ou droite pour sélectionner un nouveau runner,
// getNextUnusedRunner permet de déterminer quel est le prochain runner non sélectionné
// c'est cette fonction qui permet de ne pas pouvoir sélectionner 2 fois le même runner
// et aussi de pouvoir sauter par dessus ceux qui sont déjà sélectionnés
func getNextUnusedRunner(id int, colorScheme int, direction string) int {

	// on vérifie pour chaque client, qu'il est différent de celui que le client (id) essaye de sélectionner
	for i := 0; i < len(s.clients); i++ {
		if id != i {

			// si le client essaye de sélectionner une couleur déjà sélectionner par quelqu'un d'autre,
			// on appelle récursivement la fonction avec la couleur d'avant ou d'après en fonction de la direction
			// jusqu'à ce qu'on trouve une couleur qui n'est utilisé par personne
			if s.clients[i].colorScheme == colorScheme {
				if direction == "left" {
					return getNextUnusedRunner(id, (colorScheme + 7) % 8, direction)
				} else {
					return getNextUnusedRunner(id, (colorScheme + 1) % 8, direction)
				}
			}
		}
	}

	return colorScheme
}


// envoie le message donné au client correspondant au writer donné
func WriteToClient(writer *bufio.Writer, message string) {
	writer.WriteString(message + "|\n")
	writer.Flush()
	// log.Println("wrote : " + message + "|\n")
}


// envoie le message donné à tous les clients
func WriteToAllClients(message string) {
	for i := 0; i < len(s.clients); i++ {
		WriteToClient(s.clients[i].writer, message)
	}
}


// envoie le message donné à tous les clients sauf celui don't l'id est égal à l'idToNotWrite donné
func WriteToAllClientsExceptOne(message string, idToNotWrite int) {
	for i := 0; i < len(s.clients); i++ {

		// on envoie le message que si l'id du client est différent de celui auquel on ne doit pas écrire
		if idToNotWrite != s.clients[i].id {
			WriteToClient(s.clients[i].writer, message)
		}
	}
}


// lit les messages envoyés par le client
func ReadFromClient(client Client) {
	var reader *bufio.Reader
	reader = bufio.NewReader(client.conn)

	for {
		// on lit les messages envoyés par le client
		msg, err := reader.ReadString('\n')

		if err != nil {
			// log.Println("Erreur : ", err)
			return
		}

		// si on a reçu un message
		if msg != "" {

			// on découpe le string qui est sous la forme commance|argument1|argument2|...|\n
			splitString := strings.Split(msg, "|")

			if splitString != nil && len(splitString) > 0 {

				// la première case du tableau nous permet de savoir où nous sommes rendus dans le déroulement du jeu
				// et ce que nous sommes ainsi censés faire
				switch splitString[0] {

				// si le joueur a appuyé sur les flèces gauche ou droite pendant la sélection des runners alors on détermine
				// quel runner il devrait sélectionner (il doit sauter par-dessus ceux qui sont déjà sélectionnés)
				case "playerChangedRunner":
					var nextUnusedRunner int

					// on détermine la prochaine couleur de runner qui n'est sélectionné par personne (en fonction quelle flèce a été appuyé (gauche ou droite))
					if splitString[1] == "left" {
						nextUnusedRunner = getNextUnusedRunner(client.id, (s.clients[client.id].colorScheme + 7) % 8, "left")
					} else {
						nextUnusedRunner = getNextUnusedRunner(client.id, (s.clients[client.id].colorScheme + 1) % 8, "right")
					}

					// on modifie la couleur sélectionné par le client
					s.clients[client.id].colorScheme = nextUnusedRunner

					// on dit à tous les clients quelle couleur a été sélectionné par le client
					WriteToAllClients("playerChangedRunner|" + strconv.Itoa(client.id) + "|" + strconv.Itoa(nextUnusedRunner))


				// valide ou annule la sélection d'un runner
				case "playerSelectedRunner":

					// si le joueur a déjà sélectionné son runner, on annule la sélection.
					if s.clients[client.id].colorSelected {
						s.clients[client.id].colorSelected = false

						s.nbOfSelectedRunners--

					// si le joueur n'a pas encore sélectionné son runner, on le sélectionne
					} else {
						// on appelle une fois la fonction getNextUnusedRunner pour s'assurer que le client n'a pas la même sélection qu'un autre (non utilisé finalement)
						// s.clients[client.id].colorScheme = getNextUnusedRunner(client.id, s.clients[client.id].colorScheme, "right")

						// on valide la sélection
						s.clients[client.id].colorSelected = true

						// on augmente de 1 le nombre de joueurs qui ont validé la sélection de leur runner
						s.nbOfSelectedRunners++

						// si tous les joueurs ont validé leur runner, on démarre le compte à rebours
						if s.nbOfSelectedRunners == 4 {
							WriteToAllClients("startCountdown")
						}
					}

					// on envoie à tous les clients le runner qui a été validé par le client
					WriteToAllClients("playerSelectedRunner|" + strconv.Itoa(client.id) + "|" + strconv.Itoa(s.clients[client.id].colorScheme))


				// quand un runner se déplace, on met à jour sa position pour tous les clients
				case "updateRunnerPosition":
					// on envoie à tous les clients, à part celui qui nous a envoyé l'information, la position et la vitesse du runner
					WriteToAllClientsExceptOne("updateRunnerPosition|" + strconv.Itoa(client.id) + "|" + splitString[1] + "|" + splitString[2], client.id)


				// quand un runner est arrivé, on prévient les autres clients
				case "runnerArrived":

					// on envoie à tous les clients, à part celui qui nous a envoyé l'information, le temps du runner
					WriteToAllClientsExceptOne("runnerArrived|" + strconv.Itoa(client.id) + "|" + splitString[1], client.id)

					// on augmente de 1 le nombre de runners qui sont arrivés
					s.nbOfArrivedRunners++

					// si tous les runners sont arrivés
					if s.nbOfArrivedRunners == 4 {
						// on réinitialise la variable pour la prochaine course
						s.nbOfArrivedRunners = 0

						// on dit à tous les clients d'afficher les résultats
						WriteToAllClients("showResults")
					}

				// quand un runner est prêt à recommencer
				case "playerIsReadyToRestart":

					// on augmente de 1 le nombre de joueurs prêts à recommencer
					s.nbOfPlayersReadyToRestart++

					// on envoie à tous les clients le nombre de joueurs prêts à recommencer
					WriteToAllClients("playerIsReadyToRestart|" + strconv.Itoa(s.nbOfPlayersReadyToRestart))

					// si tous les joueurs sont prêts à recommencer
					if s.nbOfPlayersReadyToRestart == 4 {
						// on réinitialise la variable pour la prochaine course
						s.nbOfPlayersReadyToRestart = 0
					}
				}
			}
		}
	}
}


func main() {

	// on écoute les connexions sur le port 8080
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Println("listen error:", err)
		return
	}
	defer listener.Close()

	// on boucle 4 fois de sorte à avoir les 4 connexions
	for i := 0; i < 4; i++ {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("accept error:", err)
			return
		}

		// on crée un nouveau client lorsqu'il se connecte
		client := Client{
			i, // id
			conn,
			bufio.NewWriter(conn),
			0, // color scheme
			false} // color selected

		// on ajoute le client au tableau de clients de s
		s.clients = append(s.clients, client)

		// on envoie au client son id qui permet de l'identifier
		WriteToClient(client.writer, "id|"+strconv.Itoa(i))

		// on informe tous les clients du nombre de joueurs connectés
		WriteToAllClients("waitingForPlayers|" + strconv.Itoa(len(s.clients)))

		// on exécute la fonction qui permet de lire les messages envoyés par le client
		// on utilise une goroutine pour ne pas bloquer le reste de l'exécution du programme
		go ReadFromClient(client)

		defer conn.Close()
	}

	// boucle infinie pour maintenir les goroutines en vie
	for {}
}