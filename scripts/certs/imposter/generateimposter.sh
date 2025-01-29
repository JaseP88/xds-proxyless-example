################################ Imposter CA ################################
# Generates Imposter CA priv key
openssl genrsa -out imposterCA.key 4096
# Generates Imposter CA cert
openssl req -x509 -new -nodes -key imposterCA.key -sha256 -subj "/C=US/ST=MO/L=STL/O=Example CA Inc./CN=Imposter CA" -days 365 -out imposterCA.crt
# Convert to PEM
openssl x509 -in imposterCA.crt -out imposterCA.pem -outform PEM


################################ GRPC Imposter Client ################################
# Generate GRPC Imposter Client priv key
openssl genrsa -out imposter_grpcclient.key 4096
# Convert to PEM
openssl rsa -in imposter_grpcclient.key -out imposter_grpcclient_key.pem

# Generate GRPC Client conf 
cat > imposter_grpcclient.conf <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = imposter_grpcclient.key
default_md         = sha256
prompt             = no
distinguished_name = req_distinguished_name
x509_extensions    = v3_req
 
[ req_distinguished_name ]
C                  = US
ST                 = MO
L                  = STL
O                  = Example Org
OU                 = Example Org Un
CN                 = common-grpcclient
 
[ v3_req ]
keyUsage           = keyEncipherment, dataEncipherment
extendedKeyUsage   = clientAuth,serverAuth
subjectAltName     = @alt_names
 
[ alt_names ]
DNS.1              = localhost
IP.1               = 127.0.0.1
EOF

# Create a CSR for Imposter GRPC Client 
openssl req -new -key imposter_grpcclient.key -out imposter_grpcclient.csr -config imposter_grpcclient.conf
# Create a cert for Imposter GRPC Client
openssl x509 -req -in imposter_grpcclient.csr -CA imposterCA.crt -CAkey imposterCA.key -CAcreateserial -out imposter_grpcclient.crt -days 365 -sha256 -extensions v3_req -extfile imposter_grpcclient.conf
# Convert to PEM
openssl x509 -in imposter_grpcclient.crt -out imposter_grpcclient.pem -outform PEM



################################ GRPC Imposter Server ################################
# Generate GRPC Imposter Server priv key
openssl genrsa -out imposter_grpcserver.key 4096
# Convert to PEM
openssl rsa -in imposter_grpcserver.key -out imposter_grpcserver_key.pem

# Generate GRPC Server conf 
cat > imposter_grpcserver.conf <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = imposter_grpcserver.key
default_md         = sha256
prompt             = no
distinguished_name = req_distinguished_name
x509_extensions    = v3_req
 
[ req_distinguished_name ]
C                  = US
ST                 = MO
L                  = STL
O                  = Example Org
OU                 = Example Org Un
CN                 = common-grpcserver
 
[ v3_req ]
keyUsage           = keyEncipherment, dataEncipherment
extendedKeyUsage   = clientAuth,serverAuth
subjectAltName     = @alt_names
 
[ alt_names ]
DNS.1              = localhost
IP.1               = 127.0.0.1
EOF

# Create a CSR for Imposter GRPC Server 
openssl req -new -key imposter_grpcserver.key -out imposter_grpcserver.csr -config imposter_grpcserver.conf
# Create a cert for Imposter GRPC Server
openssl x509 -req -in imposter_grpcserver.csr -CA imposterCA.crt -CAkey imposterCA.key -CAcreateserial -out imposter_grpcserver.crt -days 365 -sha256 -extensions v3_req -extfile imposter_grpcserver.conf
# Convert to PEM
openssl x509 -in imposter_grpcserver.crt -out imposter_grpcserver.pem -outform PEM

