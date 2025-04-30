# Rapport détaillé ;) fait par Hugo, Nathan et Margo
## Lien vers le dépôt

## Questions
### 1. En suivant les instructions du tp, nous allons :
- créer un fichier docker-compose.yml
- ajouter service db s'appuyant sur l'image Docker postgres:latest
```yaml
services:
  db:
    image: postgres:latest
```

### 2. Pour créer la base de données city_api, 
nous faisons la requete `create table` dans le fichier `database.sql` en mettant des attributs qu'on a besoin avec leur type et contrainte.

### 3. Pour créer un service web : 
Nous allons créer des fichiers:
- `main.go` : fichier principal où le serveur est configuré et démarré.
- `city.go` : fichier contenant le modèle de la ville avec ses attributs.

Le fichier `main.go` contient : 
- Importation des bibliothèques
- Variables globales et configuration
- Gestion des routes (fonctions CityHandler et HealthHandler)
- Fonctions d'assistance (getEnvOrDefault et requireEnv permettent de récupérer les variables d'environnement)

Le fichier `city.go` contient le modèle `City`, qui représente les informations géographiques et administratives d'une ville française.

### 4. Tests
Les tests demandés dans le sujet de TP sont dans le fichier `main_test.go`.

### 5. Documentation du `Dockerfile`

Le Dockerfile est structuré en deux étapes : build et production, permettant de créer une image Docker optimisée pour une application web Go.

#### A. Build (Compilation de l'application)
L’image de base `golang:1.16-alpine` est utilisée pour compiler l’application en un exécutable optimisé.
* Configuration : Définition du répertoire de travail /app.
* Gestion des dépendances : Copie des fichiers go.mod et go.sum, puis installation des dépendances avec go mod download.
* Compilation : Copie du code source suivi de la génération de l’exécutable city-api via go build.

#### B. Production (Exécution de l’application)
L’image alpine:latest est choisie pour sa légèreté.
* Configuration : Définition du port par défaut via `ENV CITY_API_PORT=2022` et exposition avec `EXPOSE ${CITY_API_PORT}`.
* Exécution : L’exécutable est copié depuis la phase de build (`COPY --from=build`), et le conteneur est lancé avec CMD `["./city-api"]`.

### 6. Workflow

Le fichier action.yml définit un workflow GitHub Actions. Dans lequel on va exécuter un job nommé `lint` sur le code Go à chaque push sur les branches `main` et `feature/*`.

#### Configuration et Déclenchement
* Événement déclencheur : Le workflow est activé sur chaque push aux branches ciblées.
* Environnement : Il s'exécute sur une machine virtuelle Ubuntu (ubuntu-latest).

Le job lint contient plusieurs étapes :
* Récupération du code source
* Télécharge le dépôt GitHub dans l’environnement de la machine virtuelle
* Installation de Go
* Configure Go en installant la version 1.23, essentielle pour compiler et exécuter le projet
* Installation de `golangci-lint`
* Télécharge et installe golangci-lint, l’outil utilisé pour analyser le code Go
* Exécution du linter
* Lance le linter sur le code source

### 7. 
Dans le même fichier, `actions.yml`, le job `test` définit un workflow GitHub Actions permettant d’exécuter automatiquement les tests unitaires du projet à chaque push sur les branches `main` et `feature/*`.

Le job test suit plusieurs étapes :
* Récupération du code source avec `actions/checkout@v2`.
* Installation de Go (`setup-go@v2`), avec la version spécifiée.
* Exécution des tests unitaires avec `go test -v` et affichage des résultats.

### 8. 
Toujours en continuant dans le fichier `actions.yml`, on définit un job build chargé de construire l'image Docker du projet à chaque push sur les branches `main` et `feature/*`.

Le job build comprend les étapes suivantes : 
* Récupération du code source (actions/checkout@v2).
* Authentification à Docker Hub via GitHub Secrets.
* Construction de l'image Docker avec la commande :
```bash
docker build -t city-api:latest .
```
* Création d'un tag docker avec la commande : 
``` bash
docker tag city-api '${{ secrets.REGISTRY_URL }}/city-api:latest'
```

### 9.
On ajoute une nouvel étape dans le `build` pour le push de l'image docker : 
```yml 
- name: Push Docker image
  run: docker push -a '${{ secrets.REGISTRY_URL }}/city-api'
```

### 10. 
Le fichier `release.yaml` définit un workflow GitHub Actions permettant de builder et publier l'image Docker de l'application lorsqu'un tag de version (format vX.X.X) est poussé sur le dépôt.

#### Publication de l'image Docker
Le job de release se déroule en plusieurs étapes :
* Checkout du code (actions/checkout@v2).
* Connexion à Docker Hub (authentification via secrets GitHub Actions).
* Build de l'image Docker en utilisant le tag poussé (extraction de la version depuis le tag Git).
* Push de l'image Docker sur Docker Hub, avec un tag correspondant à la version du tag Git (`city-api:X.X.X`).

### 11. Scan de vulnérabilités dans la CI
Pour scanner les vulnérabilités, nous allons utiliser Trivy. C'est un outil pour vérifier et analyser la sécurité des images Docker, des systèmes d'exploitation et des bibliothèques.

Dans le fichier de `release.yaml`, nous ajoutons un job `Run Trivy vulnerability scanner`. Ceci analysera l'image Docker`city-api`.
Le paramètre `vuln-type: 'os,library'` spécifie les types de vulnérabilités à scanner. Ici, Trivy analysera les vulnérabilités au niveau du système d'exploitation (OS) ainsi que celles présentes dans les bibliothèques utilisées par l'image.
On filtre également les vulnérabilités en fonction de leur gravité. Seules les vulnérabilités de niveau CRITICAL ou HIGH seront remontées dans le rapport et affecteront le statut du pipeline(`severity: 'CRITICAL,HIGH'`).

### 12. Installer k8s
`curl -sfL https://get.k3s.io | sh -`

[Source](https://docs.k3s.io/quick-start)

### 13.
#### Avant écrire un Helm chart, nous avons besoin : 
- Installer kubectl
    - Télécharger la dernière version
`curl -LO https://dl.k8s.io/release/$(curl -Ls https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl`
    - Rendez le binaire kubectl exécutable.
`chmod +x ./kubectl`
    - Déplacez le binaire dans votre PATH
`sudo mv ./kubectl /usr/local/bin/kubectl`
    - Testez pour vous assurer que la version que vous avez installée est à jour:
`kubectl version --client`

[Source](https://kubernetes.io/fr/docs/tasks/tools/install-kubectl/)

- Installer Helm [Source](https://helm.sh/docs/intro/install/)
Comme on prefère l'installation avec package manager et on est sur debian, donc nous allons faire ça :
```bash
curl https://baltocdn.com/helm/signing.asc | gpg --dearmor | sudo tee /usr/share/keyrings/helm.gpg > /dev/null
sudo apt-get install apt-transport-https --yes
echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/helm.gpg] https://baltocdn.com/helm/stable/debian/ all main" | sudo tee /etc/apt/sources.list.d/helm-stable-debian.list
sudo apt-get update
sudo apt-get install helm
``` 
#### Ecriture du Helm chart
Le chart contient les éléments suivants :
`Chart.yaml` :
* Définit les métadonnées (nom, description, version) et les dépendances (postgresql).

`values.yaml` :
* Paramétrage des déploiements (image Docker, nombre de réplicas, port d'écoute, configuration de la base de données, variables d'environnement).

Templates Kubernetes :

* `deployment.yaml` : Déploie l’application, configure les variables d’environnement nécessaires.
* `service.yaml` : Expose l’application sur le réseau interne Kubernetes.

### 14. 
```bash
helm install city-api chart/
``` 

### 15.
L'intégration Prometheus permet de collecter et d'exposer des métriques sur l'état et l'utilisation de l'application Go.

Ce code enregistre une métrique Prometheus : 
```go 
var (
 	requestCount = prometheus.NewCounterVec(
 		prometheus.CounterOpts{
 			Name: "http_requests_total",
 			Help: "Number of HTTP requests received",
 		},
 		[]string{"path"},
    )
)
main() {
    prometheus.MustRegister(requestCount)
    ...
    http.Handle("/metrics", promhttp.Handler())
```
Expose les métriques sur l’URL `/metrics` sous un format compréhensible par Prometheus.

### 16.
Le fichier `docker-compose.yml` intègre Prometheus pour récupérer ces métriques :
```yml 
 prometheus:
 image: prom/prometheus:latest
 volumes:
   - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml
 ports:
   - "9090:9090"
```
Le volume monté (`prometheus.yml`) configure Prometheus pour « scraper » (récupérer périodiquement) les métriques de l'application exposées à l’adresse `/metrics`.

### 17.
L'intégration de Grafana dans l'infrastructure Docker permet de visualiser facilement les métriques collectées par Prometheus, offrant un tableau de bord clair et interactif pour monitorer l'application.

Le service Grafana est ajouté au fichier docker-compose.yml :
```yml
    grafana:
        image: grafana/grafana:latest
        ports:
        - "3000:3000"
        environment:
        - GF_SECURITY_ADMIN_USER=grafana
        - GF_SECURITY_ADMIN_PASSWORD=grafana
        volumes:
        - grafana_data:/var/lib/grafana
        depends_on:
        - prometheus

volumes:
   postgres_data:
   grafana_data:
```

Création du dashboard Grafana :
Pour créer un dashboard :
* Accédez à Grafana sur http://localhost:3000.
* Connectez-vous avec les identifiants configurés (grafana / grafana).
* Créer un dashboard, choissisez une métrique prometheus et valider