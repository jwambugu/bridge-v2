sleep 5 &&
    curl \
        --header "X-Vault-Token: secret" \
        --request POST \
        --data '
        {
            "data": {
                "database:dsn": "postgres://postgres:secret@localhost:5432/bridge?sslmode=disable",
                "jwt_key": "9I0tDC5S789bA6sg&l5c88p@@!i18W5v"
            }
        }

        ' \
        http://vault:8200/v1/secret/data/bridge
