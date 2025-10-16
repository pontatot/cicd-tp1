# Rapport ARGOCD

## Introduction ArgoCD

ArgoCD est un outil de déploiement automatisé à partir de stockage de source. Il permet de déployer une chart helm sur un cluster kubernetes. Pouvant être déployé sur kubernetes, cela permet de sécuriser les deploiements, car aucune clé d'accès n'a besoin d'exister dans les pipelines de CI/CD.

Il fonctionne par synchronisation en vérifiant l'état demandé par la chart helm des sources et l'état actuel déployé dans le cluster. Contrairement à ce que nous pourrions penser, il n'utilise pas directement Helm pour le déploiement, il créait les ressources manuellement avec l'api kube étant lui-même un opérateur kubernetes.

## Interprétation du sujet

à partir d'une application nécessitant une base de données (City API dans notre cas), modifier le déploiement pour utiliser ArgoCD avec comme contraintes :

1) Un environnement de production stable mis à jour lors des releases (commit sur main dans notre cas)
2) Des environnement de dev/tests créée à la volée lors de création de Pull Request(PR) sur GitHub.
3) La base de données des environnements de dev/tests doit être une copie de celle de production afin d'avoir des données, mais ne doit pas impacter celle de production si modifiée.
4) Les environnements de dev/tests doivent être nettoyés(supprimés) après fermeture de la PR.
5) Les migrations de base de données doivent être exécutées dans tous les environnements.

## Réalisation

### 1) Environnement PROD

Feature de base ArgoCD, réalisée dans [`application-production.yaml`](./argocd/application-production.yaml)

Pour les mises à jour, la CI build les images docker avec en tag le sha du commit. À chaque commit de main, la [CI](./.github/workflows/build.yml) met à jour le tag de l'image docker dans la chart HELM ce qui force le redéploiement ArgoCD avec la nouvelle image.

### 2) Environnement DEV

Après recherche dans la documentation ArgoCD, j'ai trouvé une page à propos des "[Pull Request Generator](https://argo-cd.readthedocs.io/en/latest/operator-manual/applicationset/Generators-Pull-Request/)". C'est une feature de ArgoCD qui permet de découvrire les Pull Requests (de github dans notre cas) et de déployer une instance d'application templatisée à partir d'informations sur celle-ci.

Réalisé dans [`applicationset-pr-environments.yaml`](./argocd/applicationset-pr-environments.yaml), ArgoCD monitore notre repo github et crée un nouvel environnement à partir de chaque PR.

Similairement à l'environnement de PROD, ici aussi nous utilisons les SHA des commits git pour le tag des images afin de garantir que l'application déployée soit la dernière possible même en cas de mise à jour de la PR.

### 3) Base de données copiée

Dans l'optique d'avoir des données cohérentes dans les environnements de developpements/tests, une copie de la base de données de production est effectuées. Pour cela, dans le déploiement de notre application([`deployment.yaml`](./chart/templates/deployment.yaml)) un init-container sers à attendre que notre BDD locale soit prête puis un autre viens injecter les données de celle de PROD dedans grâce à l'utilitaire `pg_dump`.

La base de données locale étant une copie de celle de production, elle ne l'affecte donc aucunement.

### 4) Nettoyage des environnements

Le "Pull Request Generator" d'ArgoCD mentionné précédemment nous permet automatiquement de nettoyer le namespace grâce au champ `finalizers`, cependant le namespace créé ne sera pas supprimé. J'ai donc ajouté une fonctionnalité à un job existant [`automated-secret-management.yaml`](./argocd/automated-secret-management.yaml) qui s'occupe de créer les secrets pour les imagePullSecrets de mes environnements. Il arrive maintenant aussi à détecter les namespaces non utilisé d'ArgoCD pour les supprimer.

### 5) Migrations

Notre application de base ne supportait pas les migrations de base de données

## Déploiement

### Prérequis

- Cluster kubernetes avec permissions admin
- ArgoCD avec permissions admin
- Github token
    - `repo` permissions de lecture des repo privés
    - `read:packages` permissions pour pull l'image docker de ghcr
    - `read:org` pour les repos d'organisations

### Secrets Github

```SH
TOKEN="YOUR_GITHUB_TOKEN"

# GitHub token pour les repos privés
kubectl create secret generic github-token \
  --from-literal=token="$TOKEN" \
  --namespace=argocd

# credentials GitHub pour pull les images ghcr
kubectl create secret generic github-credentials \
  --from-literal=username="pontatot" \
  --from-literal=token="$TOKEN" \
  --namespace=argocd
```

### Setup ArgoCD

```SH
# Cron Job pour setup les secrets ImagePullSecrets et nettoyage des namespaces
kubectl apply -f argocd/automated-secret-management.yaml

# Environement production
kubectl apply -f argocd/application-production.yaml

# Environement pour les PR
kubectl apply -f argocd/applicationset-pr-environments.yaml
```
