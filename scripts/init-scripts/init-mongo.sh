#!/bin/bash

# Conecta-se ao MongoDB
mongo <<EOF
use vilow

db.createCollection('default')

exit
EOF