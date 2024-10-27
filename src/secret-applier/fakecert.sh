#!/bin/bash

# Configura i parametri di default
DOMAIN=""
DAYS=365
KEY_SIZE=2048
BASE_DIR="/etc/letsencrypt/live"

# Funzione per mostrare l'help
usage() {
    echo "Uso: $0 -d dominio [-t giorni] [-s dimensione_key]"
    echo
    echo " -d DOMAIN       Dominio per il certificato (obbligatorio)"
    echo " -t DAYS         Giorni di validità (default: 365)"
    echo " -s KEY_SIZE     Dimensione chiave in bit (default: 2048)"
    echo
}

# Parsing degli argomenti
while getopts "d:t:s:h" option; do
    case "$option" in
        d) DOMAIN=$OPTARG ;;
        t) DAYS=$OPTARG ;;
        s) KEY_SIZE=$OPTARG ;;
        h|*) usage; exit 1 ;;
    esac
done

# Controllo se il dominio è stato fornito
if [[ -z "$DOMAIN" ]]; then
    echo "Errore: dominio non specificato!"
    usage
    exit 1
fi

# Definisci il percorso per i certificati del dominio
DOMAIN_DIR="$BASE_DIR/$DOMAIN"

# Crea la directory se non esiste
mkdir -p "$DOMAIN_DIR"

# Genera la chiave privata
openssl genpkey -algorithm RSA -out "$DOMAIN_DIR/privkey.pem" -pkeyopt rsa_keygen_bits:$KEY_SIZE

# Genera la richiesta di certificato (CSR)
openssl req -new -key "$DOMAIN_DIR/privkey.pem" -out "$DOMAIN_DIR/$DOMAIN.csr" -subj "/CN=$DOMAIN"

# Genera un certificato autofirmato
openssl x509 -req -in "$DOMAIN_DIR/$DOMAIN.csr" -signkey "$DOMAIN_DIR/privkey.pem" \
    -out "$DOMAIN_DIR/cert.pem" -days "$DAYS"

# Crea un file di catena fittizio
cp "$DOMAIN_DIR/cert.pem" "$DOMAIN_DIR/chain.pem"

# Crea un file fullchain combinando il certificato e la catena fittizia
cat "$DOMAIN_DIR/cert.pem" "$DOMAIN_DIR/chain.pem" > "$DOMAIN_DIR/fullchain.pem"

echo "Certificato generato nella cartella di Certbot:"
echo "Percorso chiave privata: $DOMAIN_DIR/privkey.pem"
echo "Percorso certificato: $DOMAIN_DIR/cert.pem"
echo "Percorso catena fittizia: $DOMAIN_DIR/chain.pem"
echo "Percorso fullchain: $DOMAIN_DIR/fullchain.pem"