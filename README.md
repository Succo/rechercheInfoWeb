# A repo with exercices for the "Recherche information web" course.

Uses [plot](https://github.com/gonum/plot) to draw the required plots and [porter2](https://github.com/surgebase/porter2) for stemming.
[Snappy](https://google.github.io/snappy/) is used to improve the size of the encoded indexes.

A working version of the code is available [here](https://riw.succo.fr).


# Notice d'intallation

Le projet est disponible sur [github](https://github.com/Succo/rechercheInfoWeb).

Mon projet de RIW a été réalisé en [golang](https://golang.org/).
Pour pouvoir lancer mon projet il faut le compilateur trouvable [ici](https://golang.org/dl/).

Un particularité de go est l'utilisation du GOPATH qui est la racine d'un dossier ou seront tous les programmes et librairies liè à go.
Par défaut la valeur de $GOPATH est $HOME/go, il est possible de choisir un autre chemin en modifiant la variable d'environnement.

Une fois go installé il est possible d'installer directement le dossier avec tous le code avec `go get gitub.com/Succo/rechercheInfoWeb`.
Cela devrait installer toutes les dépendances.
Il est aussi possible d'installer les dépendances une par une avec
```
go get github.com/gonum/plot
go get github.com/gonum/floats
go get github.com/surgebase/porter2
go get github.com/golang/snappy
```

Pour utiliser le programme il faut le compiler en lancant `go install` dans la racine du dossier `$GOPATH/src/github.com/Succo/rechercheInfoWeb`.
Le binaire produit sera `$GOPATH/bin/rechercheInfoWeb`.

Pour lancer le binaire il faut

1. Avoir le dossier cacm dans `data/CACM`
2. Avoir le dossier CS276 dans `data/CS276/pa1-data`
3. Avoir un dossier graphs en indexes (idéalement vide pour ne pas prendre le risque de perdre des données)
4. Avoir le dossier template sous la racine de la ou le programme est executé

Cela correspond à la configuration de ce dossier à condition de lancer
```
mkdir graphs indexes
wget http://web.stanford.edu/class/cs276/pa/pa1-data.zip -P data/CS276
unzip data/CS276/pa1-data.zip
```

Dans ces conditions la commande `rechercheInfoWeb -index` devrait génerer les index et lancer le serveur, `rechercheInfoWeb` seul relance le serveur en chargeant des index existant.
Il est possible d'ajouter l'argument `-precall` à ces deux commandes pour avoir les graphes de précision rappel.

Dans tous les cas lorsque le serveur est lancé il est possible d'y accèder [http://localhost:8080](http://localhost:8080).
L'interface permet de lancer des requètes sur les différents corpus avec différentes option.

Dans tous les cas le même programme est disponible en ligne à [https://riw.succo.fr](https://riw.succo.fr).

+ Le détails de la structure du programme est disponible à [https://riw.succo.fr/archi](https://riw.succo.fr/archi).
+ Des mesures de performances sont indiqué ici [https://riw.succo.fr/perf](https://riw.succo.fr/perf).
+ Et des percentiles sur les temps moyen des requètes sont là [https://riw.succo.fr/percentile](https://riw.succo.fr/percentile)  mais assez imprécis en raison du faible nombre de requète.
+ Les graphes de précision rappel pour l'ensembles des requètes (ayant donné des résultats) de CACM sont [https://riw.succo.fr/qrels](https://riw.succo.fr/qrels) et le graphe moyenné avec les valeurs de MAPS est [https://riw.succo.fr/precall](https://riw.succo.fr/precall).

Toutes ces pages sont aussi accessible localement à l'adresse donné ci dessus tant que le serveur tourne.
