# TRÅDFRI semester 2 challenge
Dit is mijn semester 2 challenge om te communiceren met de TRÅDFRI hub door een Golang programma. Doormiddel van een Golang programma communiceer je met de Ikea TRÅDFRI hub. Door te communiceren met de Ikea hub kan je dan de lampen van de set aansturen. Het hoofdprogramma draait op de Raspberry. Ook is er een server waar de acties worden opgeslagen. Daar is een apart programma voor die op de server staat en luistert of het hoofdprogramma wat verstuurd. Voor de beveiliging maak ik gebruik van DTLS (Datagram Transport Layer Security) in de code. Op die manier kunnen gebruikers het product veilig gebruiken. Afluisteren, knoeien of vervalsing van berichten worden met DTLS dus voorkomen.  

Het hoofdprogramma draait dus op de Raspberry en heet ```tradfri.go```. In dat programma wordt de data verzameld aan de hand van de gebruikers input en vervolgens in JSON gezet. Deze data wordt dan verstuurd naar een server waar de data in een database gezet word. Dat gaat met het programma ```ontvangData.go```.




# Raspberry
Op de Raspberry moet je Golang installeren. Dat gaat het makkelijkste via https://go.dev/dl/. Je neemt de koppeling van de versie die je wil. Dan voer je in je command line de volgende opdrachten in (zorg wel voor de nieuwste stabiele versie):


```wget https://go.dev/dl/go1.18.2.linux-armv6l.tar.gz```
```sudo tar -C /usr/local -xvf go1.18.2.linux-armv6l.tar.gz```
```cat >> ~/.bashrc << 'EOF'```
```export GOPATH=$HOME/go```
```export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin```
```EOF```
```source ~/.bashrc```
```go version```


Voor verdere hulp kan je hier kijken: https://gist.github.com/simoncos/49463a8b781d63b5fb8a3b666e566bb5. 


In de code staat ```const Endpoint = "http://192.168.44.147:5000/reading"```. Dit moet je aanpassen naar de voor jou juiste endpoint. Dat is dus waar je de data heen stuurt om je server hem te laten ontvangen. 



**Nodige libraries voor de Raspberry**

Om de libraries op te halen gebruik je de volgende commando's in je command line op de Raspberry. Zorg dat Golang al geïnstalleerd is. 

```go get github.com/eriklupander/dtls```
```go get github.com/eriklupander/tradfri-go/router```
```go get github.com/eriklupander/tradfri-go/tradfri```
```go get github.com/sirupsen/logrus```
```go get github.com/spf13/pflag```
```go get github.com/spf13/viper```


# Server
Op de server zet je het programma ```ontvangData.go```. Ook op de server moet je zorgen dat je Golang geïnstalleerd hebt. Dat gaat het makkelijkste via https://go.dev/dl/. Je neemt de koppeling van de versie die je wil. 

Dan voer je in je command line de volgende opdrachten in (zorg wel voor de nieuwste stabiele versie): 

```wget https://go.dev/dl/go1.18.2.linux-amd64.tar.gz```
```sudo tar -C /usr/local -xvf go1.18.2.linux-amd64.tar.gz```
```cat >> ~/.bashrc << 'EOF'```
```export GOPATH=$HOME/go```
```export PATH=/usr/local/go/bin:$PATH:$GOPATH/bin```
```EOF```
```source ~/.bashrc```
```go version```

Voor verdere hulp kan je hier kijken: https://gist.github.com/simoncos/49463a8b781d63b5fb8a3b666e566bb5. 

Ook moet je zorgen dat je gccgo op de server hebt staan, omdat je dat nodig hebt om gebruik te maken van de libraries die in de code zitten. 

```sudo apt install gccgo```

Ook is het belangrijk dat je SQLite heb geïnstalleerd, omdat daar de code dat gebruikt. Je hoeft geen database aan te maken of iets in te stellen verder. Dat gaat allemaal automatisch in de code. 

```sudo apt install sqlite3```



**Nodige libraries voor op de server**

Om de libraries op te halen gebruik je de volgende commando's in je command line op de Raspberry. Zorg dat Golang al geïnstalleerd is. 

```go get github.com/gin-gonic/gin```

```go get github.com/mattn/go-sqlite3```


