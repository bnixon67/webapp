openssl req -newkey rsa:4096 -x509 -sha512 -days 365 -nodes -out cert/cert.pem -keyout cert/key.pem -subj /CN=localhost
