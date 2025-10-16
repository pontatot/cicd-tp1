# Rapport ARGOCD

## Introduction ArgoCD

ArgoCD est une outil de deploiement automatisé à partir de stockage de source. Il permet de deployer une chart helm sur un cluster kubernetes. Pouvant être deployé sur kubernetes, cela permet de sécuriser les deploiement car aucune clé d'accés n'a besoin d'exister dans les pipelines de CI/CD.

Il fonctionne par synchronisation en verifiant l'état demandé par la chart helm des sources et l'état actuel deployé dans le cluster. Contrairement à ce que nous pourrions penser, il n'utilise pas directement Helm pour le deployement, il créé les ressources manuellement avec l'api kube étant lui-même un opérateur kubernetes.

## Interpretation du sujet

à partir d'une application nécessitant une base de données (City API dans notre cas), modifier le deploiement pour utiliser ArgoCD avec comme contraintes:

1) Un environnement de production stable mis à jour lors des releases (commit sur main dans notre cas)
2) Des environnement de dev/tests créée à la volé lors de création de Pull Request(PR) sur GitHub.
3) La base de données des environnements de dev/tests doit être une copie de celle de production afin d'avoir des données, mais ne doit pas impacter celle de production si mofifiée.
4) Les environnements de dev/tests doivent être nettoyés(supprimés) après fermeture de la PR.
5) Les migrations de base de données doivent être executés dans tous les environnements.

## Réalisation

### 1) Environnement PROD

Feature de base ArgoCD, réalisée dans [`application-production.yaml`](./argocd/application-production.yaml)

Pour les mises à jour, la CI build les images docker avec en tag le sha du commit. À chaque commit de main, la [CI](./.github/workflows/build.yml) met à jour le tag de l'image docker dans la chart HELM ce qui force le redeploiement ArgoCD avec la nouvelle image.

### 2) Environnement DEV

Après recherche dans la documentation ArgoCD, j'ai trouvé une page à propos des "[Pull Request Generator](https://argo-cd.readthedocs.io/en/latest/operator-manual/applicationset/Generators-Pull-Request/)". C'est une feature de ArgoCD qui permet de découvrire les Pull Requests (de github dans notre cas) et de deployer une instance d'application templatisée à partir d'informations sur celle-ci.

Réalisé dans [`applicationset-pr-environments.yaml`](./argocd/applicationset-pr-environments.yaml), ArgoCD monitore notre repo github et créée un nouveau environnement à partir de chaque PR.

Similairement à l'environnement de PROD, ici aussi nous utilisons les SHA des commit git pour le tag des images afin de garantir que l'application deployée soit la dernière possible même en cas de mise à jour de la PR.

### 3) Base de données copiée

Dans l'optique d'avoir des données cohérentes dans les environnements de developpements/tests, une copie de la base de données de production est effectuées. Pour cela, dans le deploiement de notre application([`deployment.yaml`](./chart/templates/deployment.yaml)) un init-container sers à attendre que notre BDD locale soit prète puis un autre viens injecter les données de celle de PROD dedans grâce à l'utilitaire `pg_dump`.

La base de données locale étant une copie de celle de production, elle ne l'affecte donc aucunement.

### 4) Nettoyage des environnements

Le "Pull Request Generator" d'ArgoCD mentionné précedemment nous permet automatiquement de nettoyer le namespace grace au champ `finalizers`, cependant le namespace créé ne sera pas supprimé. J'ai donc ajouter une fonctionalité à un job existant [`automated-secret-management.yaml`](./argocd/automated-secret-management.yaml) qui s'occupe de créer les secret pour les imagePullSecrets de mes environnements. Il arrive maintenant aussi à detecter les namespaces non utilisé d'ArgoCD pour les supprimer.

### 5) Migrations

Notre application de base ne supportais pas les migrations de base de données
