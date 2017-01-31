go install
rsync -avz --progress indexes fabrice@succo.fr:riw/
rsync -avz --progress templates fabrice@succo.fr:riw/
rsync -avz --progress graphs fabrice@succo.fr:riw/
# A bit wonky
rsync $GOPATH/bin/rechercheInfoWeb fabrice@succo.fr:
echo "Restart the process to finalise the deploy"
