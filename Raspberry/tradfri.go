package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/eriklupander/dtls"

	"github.com/eriklupander/tradfri-go/router"
	"github.com/eriklupander/tradfri-go/tradfri"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var configFlags = pflag.NewFlagSet("config", pflag.ExitOnError)
var commandFlags = pflag.NewFlagSet("commands", pflag.ExitOnError)
var number int
var powernumber int
var keuzelamp int
var naamlamp string
var status string

func init() {
	// alle termen die gebruikt worden
	configFlags.String("gateway_ip", "", "ip naar de gateway. Geen protocol of poort hier!")
	configFlags.String("gateway_address", "", "Adres naar de gateway. Inclusief poort hier!")
	configFlags.String("psk", "", "Pre-shared key onderaan de gateway")
	configFlags.String("client_id", "", "Een klant-ID, kan je zelf verzinnen")
	configFlags.String("loglevel", "info", "log leve. Allowed values: fatal, error, warn, info, debug, trace")

	commandFlags.Bool("server", false, "Start in server modus?")
	commandFlags.Bool("authenticate", false, "PSK-uitwisseling uitvoeren")
	commandFlags.String("get", "", "URL voor GET")
	commandFlags.String("put", "", "URL voor PUT")
	commandFlags.String("payload", "", "payload voor PUT")
	commandFlags.Int("port", 8080, "poort van de server")

	commandFlags.AddFlagSet(configFlags)
	_ = commandFlags.Parse(os.Args[1:])

	_ = viper.BindPFlags(configFlags)
	viper.AutomaticEnv()
	viper.AddConfigPath(".") // leest config.json uit
	err := viper.ReadInConfig()
	if err != nil {
		logrus.Info(err.Error())
		logrus.Info("Je zal waarschijnlijk --authenticate eerst moeten uitvoeren")
	}
	viper.RegisterAlias("pre_shared_key", "psk")
}

func logging() {

	// alle logging configureren
	levelStr := viper.GetString("loglevel")
	if levelStr == "" {
		levelStr = "info"
	}
	fmt.Printf("Using loglevel: %v\n", levelStr)
	level, err := logrus.ParseLevel(levelStr)
	if err != nil {
		fmt.Println("invalid loglevel")
		os.Exit(1)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	log.SetOutput(logrus.StandardLogger().Out)
	dtls.SetLogFunc(func(ts time.Time, level string, peer string, msg string) {
		switch level {
		case "error":
			logrus.WithField("level", level).WithField("peer", peer).Error(msg)
		case "warn":
			logrus.WithField("level", level).WithField("peer", peer).Warn(msg)
		case "info":
			logrus.WithField("level", level).WithField("peer", peer).Info(msg)
		case "debug":
			logrus.WithField("level", level).WithField("peer", peer).Debug(msg)
		}
	})
	dtls.SetLogLevel(resolveDTLSLogLevel(levelStr))

}

func main() {
	logging()

	gatewayAddress := viper.GetString("gateway_address")
	if gatewayAddress == "" {
		gatewayAddress = viper.GetString("gateway_ip") + ":5684"
	}
	psk := viper.GetString("psk")
	clientID := viper.GetString("client_id")
	serverMode, _ := commandFlags.GetBool("server")
	authenticate, _ := commandFlags.GetBool("authenticate")
	get, getErr := commandFlags.GetString("get")
	put, putErr := commandFlags.GetString("put")
	payload, _ := commandFlags.GetString("payload")
	port, _ := commandFlags.GetInt("port")

	// Omgaan met de speciale authenticatie functie
	if authenticate {
		performTokenExchange(gatewayAddress, clientID, psk)
		return
	}

	checkRequiredConfig(gatewayAddress, clientID, psk)

	// Checken op welke manier je het programma gebruikt

	if serverMode {
		// door gebruik te maken van --server kom je hier uit
		logrus.Info("Server modus start")
		logrus.Infof("REST: %d", port)

		tc := tradfri.NewTradfriClient(gatewayAddress, clientID, psk)
		// REST
		go router.SetupChi(tc, port)

		// functie voor het keuze menu wordt aangeroepen
		keuzeAansturen()

	} else {
		// client modus, hier komen de get en put opdrachten aan
		if getErr == nil && get != "" {
			resp, _ := tradfri.NewTradfriClient(gatewayAddress, clientID, psk).Get(get)
			logrus.Infof("%v", string(resp.Payload))
		} else if putErr == nil && put != "" {
			resp, _ := tradfri.NewTradfriClient(gatewayAddress, clientID, psk).Put(put, payload)
			logrus.Infof("%v", string(resp.Payload))
		} else {
			logrus.Info("Er is geen juiste input gegeven. De mogelijkheden zijn: get, put, authenticate, server")
		}
	}

}

func keuzeAansturen() {
	// als je --server modus gekozen hebt kom je hier
	// dit blijft lopen zolang server modus actief is
	for true {
		// een menu waarin je door een cijfer in te voeren door heen kan navigeren
		// en opdrachten uit kan voeren
		fmt.Print("\nWat wil je doen?", "\nLamp besturen: 1", "\nInfo opvragen: 2")
		fmt.Print("\nKies door het cijfer in te typen: ")

		fmt.Scanln(&number) // input van de gebruiker ophalen
		if number == 1 {
			fmt.Print("\nWelke lamp wil je aan of uit zetten?")
			fmt.Print("\nWoonkamer lamp: 1", "\nGarage lamp: 2")
			fmt.Print("\nKies door het cijfer te typen: ")
			fmt.Scanln(&keuzelamp) // input van de gebruiker ophalen

			if keuzelamp == 1 {
				naamlamp = "Woonkamer"
				fmt.Print("\nWil je de Woonkamer lamp aan of uit zetten?")
				fmt.Print("\nWoonkamer lamp uit: 0", "\nWoonkamer lamp aan: 1")
				fmt.Print("\nKies door het cijfer te typen: ")
				fmt.Scanln(&powernumber) // input van de gebruiker ophalen

				if powernumber == 0 {
					status = "Uit"
					fmt.Printf("De lamp gaat uit")
					// putErr == nil && put != ""
					// resp, _ := tradfri.NewTradfriClient(gatewayAddress, clientID, psk).Put(put, payload)
					// logrus.Infof("%v", string(resp.Payload))

				} else if powernumber == 1 {
					status = "Aan"
					fmt.Printf("De lamp gaat aan")

				} else {
					fmt.Printf("Dat is geen mogelijke keuze.")
				}

			} else if keuzelamp == 2 {
				naamlamp = "Garage"
				fmt.Print("\nWil je de Garage lamp aan of uit zetten?")
				fmt.Print("\nGarage lamp uit: 0", "\nGarage lamp aan: 1")
				fmt.Print("\nKies door het cijfer te typen: ")
				fmt.Scanln(&powernumber) // input van de gebruiker ophalen

				if powernumber == 0 {
					status = "Uit"
					fmt.Printf("De lamp gaat uit")

				} else if powernumber == 1 {
					status = "Aan"
					fmt.Printf("De lamp gaat aan")
				} else {
					fmt.Printf("Dat is geen mogelijke keuze.")
				}
			} else {
				fmt.Printf("Dat is geen mogelijke keuze.")
			}

		} else if number == 2 { // hier kan je informatie opvragen over een lamp of groep
			fmt.Printf("Hier komt informatie")
		} else {
			fmt.Printf("Dat is geen mogelijke keuze.")
		}
		fmt.Println("")
		sturenDatabase() // dit verwijst naar de functie voor het versturen van de data
	}
}

func sturenDatabase() { // in deze functie wordt de data verstuurd naar de server

	// de data die verstuurd gaat worden
	type reading struct {
		TimeStamp string
		Lamp      string
		Status    string
	}

	// naar de server
	const Endpoint = "http://192.168.44.147:5000/reading"

	fmt.Println("")

	// tijdstip pakken
	timeStamp := time.Now()

	newReading := reading{TimeStamp: timeStamp.Format("2006-01-02T15:04:05-0700"), Lamp: naamlamp, Status: status}

	// waarden uitprinten in de command line
	fmt.Println("Verzamelde data ", "\nMoment:", timeStamp, "\nLamp:", naamlamp, "\nStatus:", status)
	time.Sleep(2 * time.Second)
	//aanvraag aanmaken
	var requestBody, reqerr = json.Marshal(newReading)

	if reqerr != nil {
		fmt.Println("Request error: aanvraag error")
		return
	}

	// versturen naar de server
	resp, resperror := http.Post(Endpoint, "application/json", bytes.NewBuffer(requestBody))

	if resperror != nil {
		fmt.Println("Response error:", resperror)
		return
	}
	fmt.Println("Versturen naar de server...")
	time.Sleep(2 * time.Second)
	// aanvraag sluiten
	defer resp.Body.Close()
	fmt.Println("Klaar met versturen!")
	time.Sleep(2 * time.Second)

}

func checkRequiredConfig(gatewayAddress, clientID, psk string) { // checkt het config.json bestand of de command line flags
	if gatewayAddress == "" {
		fail("Niet gelukt om de gatewayAddress van de command-line flag of het config.json bestand uit te lezen")
	}
	if clientID == "" {
		fail("Niet gelukt om de clientID van de command-line flag of het config.json bestand uit te lezen")
	}
	if psk == "" {
		fail("Niet gelukt om de psk van de command-line flag of het config.json bestand uit te lezen")
	}
}

func performTokenExchange(gatewayAddress, clientID, psk string) { // zorgt voor het uitwisselen van de keys met de hub
	if len(clientID) < 1 || len(psk) < 10 {
		fail("Er moeten een clientID en een psk opgegeven worden wanneer de sleutels uitgewisseld worden")
	}

	done := make(chan bool)
	defer func() { done <- true }()
	go func() {
		select {
		case <-time.After(time.Second * 5):
			logrus.Info("(De uitwisseling van de psk kan vasthangen op \"Verbinden naar\" als de psk op de onderkant van de hub niet goed ingevoerd is)")
		case <-done:
		}
	}()

	// Client_identity wordt hier handmatig uitgevoerd voordat we de DTLS-client maken,
	// want dat is nodig bij het uitvoeren van de tokenuitwisseling
	dtlsClient := tradfri.NewTradfriClient(gatewayAddress, "Client_identity", psk)

	authToken, err := dtlsClient.AuthExchange(clientID)
	if err != nil {
		fail(err.Error())
	}
	viper.Set("client_id", clientID)
	viper.Set("gateway_address", gatewayAddress)
	viper.Set("psk", authToken.Token)
	err = viper.WriteConfigAs("config.json")
	if err != nil {
		fail(err.Error())
	}
	logrus.Info("De configuratie met de PSK en de clientID zijn opgeslagen in het config.json bestand in dezelfde map, bewaar het bestand.")
}

func fail(msg string) {
	logrus.Info(msg)
	os.Exit(1)
}

// resolveDTLSLogLevel brengt de logrus-onderdelen in kaart met de onderdelen die worden ondersteund door de DTLS library.
func resolveDTLSLogLevel(level string) string {
	switch level {
	case "fatal":
		fallthrough
	case "error":
		return dtls.LogLevelError
	case "warn":
		return dtls.LogLevelWarn
	case "info":
		return dtls.LogLevelInfo
	case "debug":
		fallthrough
	case "trace":
		return dtls.LogLevelDebug
	}
	return "info"
}
