#!/bin/bash

go install
if [ "$1" == "lite" ]
	then
		rsync -avz --progress . fabrice@succo.fr:riw/ --exclude ".git" --exclude "data" --exclude "indexes" 
	else 
		rsync -avz --progress . fabrice@succo.fr:riw/ --exclude ".git" --exclude "data"
fi
# A bit wonky
rsync -avz --progress $GOPATH/bin/rechercheInfoWeb fabrice@succo.fr:
echo "Restart the process to finalise the deploy"
