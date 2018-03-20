#!/bin/bash

go install
if [ "$1" == "lite" ]
	then
		rsync -avz --progress . fabrice@succo.fr:riw/ --exclude ".git" --exclude "data" --exclude "indexes" --exclude "vendor"
	else 
		rsync -avz --progress . fabrice@succo.fr:riw/ --exclude ".git" --exclude "data" --exclude "vendor"
fi
# A bit wonky
rsync -avz --progress $GOPATH/bin/rechercheInfoWeb fabrice@succo.fr:
echo "Restart the process to finalise the deploy"
