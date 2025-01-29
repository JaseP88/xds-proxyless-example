################################ CA ################################
# Generates CA priv key
openssl genrsa -out ca.key 4096
# Generates CA cert
openssl req -x509 -new -nodes -key ca.key -sha256 -subj "/C=US/ST=MO/L=STL/O=Example CA Inc./CN=Example Root CA" -days 365 -out ca.crt
# Convert to PEM
openssl x509 -in ca.crt -out ca.pem -outform PEM

################################ Grpc Server ################################
# Generate GRPC Server priv key
openssl genrsa -out grpcserver.key 4096
# Convert to PEM
openssl rsa -in grpcserver.key -out grpcserver_key.pem

# Generate GRPC Server conf 
cat > grpcserver.conf <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = grpcserver.key
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

# Create a CSR for GRPC Server 
openssl req -new -key grpcserver.key -out grpcserver.csr -config grpcserver.conf
# Create a cert for GRPC Server
openssl x509 -req -in grpcserver.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out grpcserver.crt -days 365 -sha256 -extensions v3_req -extfile grpcserver.conf
# Convert to PEM
openssl x509 -in grpcserver.crt -out grpcserver.pem -outform PEM



################################ GRPC Client ################################
# Generate GRPC Client priv key
openssl genrsa -out grpcclient.key 4096
# Convert to PEM
openssl rsa -in grpcclient.key -out grpcclient_key.pem

# Generate GRPC Client conf 
cat > grpcclient.conf <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = grpcclient.key
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

# Create a CSR for GRPC Client 
openssl req -new -key grpcclient.key -out grpcclient.csr -config grpcclient.conf
# Create a cert for GRPC Client
openssl x509 -req -in grpcclient.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out grpcclient.crt -days 365 -sha256 -extensions v3_req -extfile grpcclient.conf
# Convert to PEM
openssl x509 -in grpcclient.crt -out grpcclient.pem -outform PEM



################################ XDS Server ################################
# Generate XDS Server priv key
openssl genrsa -out xdsserver.key 4096
# Convert to PEM
openssl rsa -in xdsserver.key -out xdsserver_key.pem

# Generate XDS Server conf 
cat > xdsserver.conf <<EOF
[ req ]
default_bits       = 2048
default_keyfile    = xdsserver.key
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
CN                 = common-xds
 
[ v3_req ]
keyUsage           = keyEncipherment, dataEncipherment
extendedKeyUsage   = clientAuth,serverAuth
subjectAltName     = @alt_names
 
[ alt_names ]
DNS.1              = localhost
IP.1               = 127.0.0.1
EOF

# Create a CSR for XDS Server
openssl req -new -key xdsserver.key -out xdsserver.csr -config xdsserver.conf
# Create a cert for XDS Server
openssl x509 -req -in xdsserver.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out xdsserver.crt -days 365 -sha256 -extensions v3_req -extfile xdsserver.conf
# Convert to PEM
openssl x509 -in xdsserver.crt -out xdsserver.pem -outform PEM