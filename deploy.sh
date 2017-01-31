go install
rsync -avz indexes fabrice@succo.fr:riw/
rsync -avz templates fabrice@succo.fr:riw/
rsync -avz graphs fabrice@succo.fr:riw/
# A bit wonky
rsync ../../../../bin/rechercheInfoWeb fabrice@succo.fr:
echo "Restart the process to finalise the deploy"
